package socks5

import (
	"dxkite.cn/go-gateway/proto"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

// Socks5服务器
type Server struct {
	net.Conn
}

// 握手
func (conn *Server) Handshake() (err error) {
	/**
	 -->

		+----+----------+----------+
		|VER | NMETHODS | METHODS  |
		+----+----------+----------+
		| 1  |    1     | 1 to 255 |
		+----+----------+----------+
	VER 是指协议版本，因为是 socks5，所以值是 0x05
	NMETHODS 是指有多少个可以使用的方法，也就是客户端支持的认证方法，有以下值：
		0x00 NO AUTHENTICATION REQUIRED 不需要认证
		0x01 GSSAPI 参考：https://en.wikipedia.org/wiki/Generic_Security_Services_Application_Program_Interface
		0x02 USERNAME/PASSWORD 用户名密码认证
		0x03 to 0x7f IANA ASSIGNED 一般不用。INNA保留。
		0x80 to 0xfe RESERVED FOR PRIVATE METHODS 保留作私有用处。
		0xFF NO ACCEPTABLE METHODS 不接受任何方法/没有合适的方法
	METHODS 就是方法值，有多少个方法就有多少个byte
	*/
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Println("negotiation err", err)
		_ = conn.Close()
		return err
	}
	if version := buf[0]; version != Version5 {
		_ = conn.Close()
		return errProtocolVersion
	}
	nMethods := buf[1]
	methods := make([]byte, nMethods)
	if n, err := conn.Read(methods); n != int(nMethods) || err != nil {
		_ = conn.Close()
		return errProtocolMethods
	}
	/**
	<--
	+----+--------+
	|VER | METHOD |
	+----+--------+
	| 1  |   1    |
	+----+--------+
	*/
	buf[1] = 0 // NO AUTH
	_, err = conn.Write(buf)
	return err
}

// 获取链接信息
func (conn *Server) Info() (info *proto.ConnInfo, err error) {
	network, address, er := conn.handleCmd()
	if er != nil {
		return nil, er
	}
	return &proto.ConnInfo{
		Network:  network,
		Address:  address,
		Username: "",
		Password: "",
	}, nil
}

// 处理命令
func (conn *Server) handleCmd() (string, string, error) {
	/**
	-->

	+----+-----+-------+------+----------+----------+
	|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	+----+-----+-------+------+----------+----------+
	| 1  |  1  | X'00' |  1   | Variable |    2     |
	+----+-----+-------+------+----------+----------+
	VER 版本，取值是 0x05
	CMD 命令，取值如下：
		CONNECT 0x01 连接
		BIND 0x02 端口监听
		UDP ASSOCIATE 0x03 使用UDP
	RSV 是保留位，值是 0x00
	ATYP 是目标地址类型，有如下取值：
		0x01 IPv4
		0x03 域名
		0x04 IPv6
	DST.ADDR 就是目标地址
	DST.PORT 两个字节代表端口号
	*/
	header := make([]byte, 3)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		_ = conn.Close()
		return "", "", err
	}
	switch header[1] {
	case connectMethod:
		if addr, err := conn.readAddr(); err != nil {
			conn.sendReply(serverFailure)
		} else {
			return "tcp", addr, nil
		}
	case bindMethod:
		conn.sendReply(commandNotSupported)
	case udpAssociateMethod:
		conn.sendReply(commandNotSupported)
	default:
		conn.sendReply(commandNotSupported)
		_ = conn.Close()
	}
	return "", "", errCommandNotSupported
}

// 读取域
func (c *Server) readAddr() (string, error) {
	addrType := make([]byte, 1)
	if _, err := c.Read(addrType); err != nil {
		return "", err
	}

	var host string
	var port uint16

	switch addrType[0] {
	case AddrTypeIPv4:
		ipv4 := make(net.IP, net.IPv4len)
		if _, err := c.Read(ipv4); err != nil {
			return "", err
		}
		host = ipv4.String()
	case AddrTypeIPv6:
		ipv6 := make(net.IP, net.IPv6len)
		if _, err := c.Read(ipv6); err != nil {
			return "", err
		}
		host = ipv6.String()
	case AddrTypeFQDN:
		var domainLen uint8
		if err := binary.Read(c, binary.BigEndian, &domainLen); err != nil {
			return "", err
		}
		domain := make([]byte, domainLen)
		if _, err := c.Read(domain); err != nil {
			return "", err
		}
		host = string(domain)
	default:
		return "", errAddrTypeNotSupported
	}

	if err := binary.Read(c, binary.BigEndian, &port); err != nil {
		return "", err
	}
	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
	return addr, nil
}

func (c *Server) sendReply(rep uint8) {
	reply := []byte{5, rep, 0, 1}
	h, p, _ := net.SplitHostPort(c.LocalAddr().String())
	ip := net.ParseIP(h).To4()
	n, _ := strconv.Atoi(p)
	reply = append(reply, ip...)
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, uint16(n))
	reply = append(reply, port...)
	_, _ = c.Write(reply)
}

// 获取操作流
func (conn *Server) Stream() net.Conn {
	return conn
}

// 发送错误
func (conn *Server) SendError(err error) error {
	err = errServerFailure
	switch m := strings.ToLower(err.Error()); {
	case strings.Contains(m, "host"):
		err = errHostUnreachable
	case strings.Contains(m, "unreachable"):
		err = errNetworkUnreachable
	case strings.Contains(m, "refused"):
		err = errConnectionRefused
	}
	if v, ok := err.(socks5Err); ok {
		conn.sendReply(v.code)
	}
	return nil
}

// 发送连接成功
func (conn *Server) SendSuccess() error {
	conn.sendReply(succeeded)
	return nil
}

type Client struct {
	net.Conn
	Info proto.ConnInfo
}

func (conn *Client) Handshake() (err error) {
	/**
	 -->

		+----+----------+----------+
		|VER | NMETHODS | METHODS  |
		+----+----------+----------+
		| 1  |    1     | 1 to 255 |
		+----+----------+----------+
	VER 指协议版本 值是 0x05
	NMETHODS 有多少个可以使用的方法，客户端支持的认证方法，有以下值：
		0x00 NO AUTHENTICATION REQUIRED 不需要认证
		0x01 GSSAPI 参考：https://en.wikipedia.org/wiki/Generic_Security_Services_Application_Program_Interface
		0x02 USERNAME/PASSWORD 用户名密码认证
		0x03 to 0x7f IANA ASSIGNED 一般不用。INNA保留。
		0x80 to 0xfe RESERVED FOR PRIVATE METHODS 保留作私有用处。
		0xFF NO ACCEPTABLE METHODS 不接受任何方法/没有合适的方法
	METHODS 方法值，有多少个方法就有多少个byte
	*/
	if _, err = conn.Write([]byte{Version5, 2, NoAuthRequiredMethod, UsernamePasswordMethod}); err != nil {
		return
	}

	buf := make([]byte, 2)
	if _, err = io.ReadFull(conn, buf); err != nil {
		return
	}
	if buf[0] != Version5 {
		return errProtocolVersion
	}
	// Auth
	switch buf[1] {
	case NoAuthRequiredMethod:
	case UsernamePasswordMethod:
		if err = conn.basicAuth(); err != nil {
			return
		}
	}
	return nil
}

// 获取操作流
func (c *Client) Stream() net.Conn {
	return c
}

func (conn *Client) basicAuth() error {
	info := conn.Info
	/**
	  +----+------+----------+------+----------+
	  |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
	  +----+------+----------+------+----------+
	  | 1  |  1   | 1 to 255 |  1   | 1 to 255 |
	  +----+------+----------+------+----------+
	*/
	if len(info.Username) == 0 || len(info.Username) > 255 || len(info.Password) == 0 || len(info.Password) > 255 {
		return errors.New("invalid username/password")
	}
	b := []byte{UsernamePasswordVersion}
	b = append(b, byte(len(info.Username)))
	b = append(b, info.Username...)
	b = append(b, byte(len(info.Password)))
	b = append(b, info.Password...)
	if _, err := conn.Write(b); err != nil {
		return err
	}
	/**
	  +----+--------+
	  |VER | STATUS |
	  +----+--------+
	  | 1  |   1    |
	  +----+--------+
	*/
	if _, err := io.ReadFull(conn, b[:2]); err != nil {
		return err
	}
	if b[0] != UsernamePasswordVersion {
		return errors.New("invalid username/password version")
	}
	if b[1] != AuthStatusSucceeded {
		return errors.New("username/password authentication failed")
	}
	return nil
}

func (conn *Client) Connect() error {
	return conn.conn(conn.Info.Network, conn.Info.Address)
}

func (conn *Client) conn(network, address string) error {
	if network == "udp" {
		return errCommandNotSupported
	}
	/**
	+----+-----+-------+------+----------+----------+
	|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	+----+-----+-------+------+----------+----------+
	| 1  |  1  | X'00' |  1   | Variable |    2     |
	+----+-----+-------+------+----------+----------+
	VER 版本，取值是 0x05
	CMD 命令，取值如下：
		CONNECT 0x01 连接
		BIND 0x02 端口监听
		UDP ASSOCIATE 0x03 使用UDP
	RSV 是保留位，值是 0x00
	ATYP 是目标地址类型，有如下取值：
		0x01 IPv4
		0x03 域名
		0x04 IPv6
	DST.ADDR 就是目标地址
	DST.PORT 两个字节代表端口号
	*/
	b := []byte{Version5, connectMethod, 0}
	host, p, err := net.SplitHostPort(address)
	port, _ := strconv.Atoi(p)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			b = append(b, AddrTypeIPv4)
			b = append(b, ip4...)
		} else if ip6 := ip.To16(); ip6 != nil {
			b = append(b, AddrTypeIPv6)
			b = append(b, ip6...)
		} else {
			return errors.New("unknown address type")
		}
	} else {
		if len(host) > 255 {
			return errors.New("FQDN too long")
		}
		b = append(b, AddrTypeFQDN)
		b = append(b, byte(len(host)))
		b = append(b, host...)
	}
	b = append(b, byte(port>>8), byte(port))
	if _, err := conn.Write(b); err != nil {
		return err
	}

	/**
	+----+-----+-------+------+----------+----------+
	|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	+----+-----+-------+------+----------+----------+
	| 1  |  1  | X'00' |  1   | Variable |    2     |
	+----+-----+-------+------+----------+----------+
	*/
	if _, err = io.ReadFull(conn, b[:4]); err != nil {
		return err
	}
	if b[0] != Version5 {
		return errors.New("unexpected protocol version " + strconv.Itoa(int(b[0])))
	}
	if cmdErr := b[1]; cmdErr != succeeded {
		return errors.New("unknown error " + Reply(b[1]).String())
	}
	if b[2] != 0 {
		return errors.New("non-zero reserved field")
	}
	l := 2
	switch b[3] {
	case AddrTypeIPv4:
		l += net.IPv4len
	case AddrTypeIPv6:
		l += net.IPv6len
	case AddrTypeFQDN:
		if _, err := io.ReadFull(conn, b[:1]); err != nil {
			return err
		}
		l += int(b[0])
	default:
		return errors.New("unknown address type " + strconv.Itoa(int(b[3])))
	}
	if cap(b) < l {
		b = make([]byte, l)
	} else {
		b = b[:l]
	}
	if _, err = io.ReadFull(conn, b); err != nil {
		return err
	}
	return nil
}

type Socks5Identifier struct {
}

// 判断是否为HTTP协议
func (d *Socks5Identifier) Check(r io.Reader) (bool, error) {
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil {
		return false, err
	}
	return n == 1 && buf[0] == Version5, nil
}

type Config struct {
}

func (h *Config) Name() string {
	return "socks5"
}

// 创建Socks服务器
func (h *Config) Server(conn net.Conn) proto.Server {
	return &Server{
		Conn: conn,
	}
}

// 创建Socks客户端
func (h *Config) Client(conn net.Conn, info proto.ConnInfo) proto.Client {
	return &Client{
		Conn: conn,
		Info: info,
	}
}

func (h *Config) Identifier() proto.Identifier {
	return &Socks5Identifier{}
}

// 创建Socks5协议
func Proto(config *Config) proto.Proto {
	return config
}
