package mino

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
)

type RequestMessage struct {
	Network  string
	Address  string
	Username string
	Password string
}

// 编码
func (m *RequestMessage) marshal() ([]byte, error) {
	var host string
	var port int
	var ports string
	var err error
	if host, ports, err = net.SplitHostPort(m.Address); err != nil {
		return nil, err
	}
	if port, err = strconv.Atoi(ports); err != nil {
		return nil, err
	}
	// PackageInfo(b2)
	//	AddrType 3 bit
	//		- 000 IPv4
	//		- 001 IPv6
	//		- 010 HostName
	//	ProtoType 1 bit
	//      - 0 TCP
	//		- 1 UDP
	//	AuthType
	//		- 0000 No Auth
	//		- 0001 Password
	var b2 uint8 = 0
	if len(m.Username) > 0 {
		b2 |= 1
	}
	// NetworkType
	if m.Network == "udp" {
		b2 |= 1 << 4
	}
	// IP or Host
	h := false
	ip := net.ParseIP(host)
	if ip != nil {
		if v := ip.To4(); v != nil {
			ip = v
		} else {
			b2 |= 1 << 5
		}
	} else {
		b2 |= 1 << 6
		h = true
	}
	buf := make([]byte, 0, 2+len(host)+2+2+len(m.Username)+len(m.Password))
	buf = append(buf, Version2, b2)
	if h {
		buf = append(buf, byte(len(host)))
		buf = append(buf, host...)
	} else {
		buf = append(buf, ip...)
	}
	// Port
	var pb [2]byte
	binary.BigEndian.PutUint16(pb[:], uint16(port))
	buf = append(buf, pb[:]...)
	// AuthMessage
	if len(m.Username) > 0 {
		buf = append(buf, byte(len(m.Username)))
		buf = append(buf, byte(len(m.Password)))
		buf = append(buf, m.Username...)
		buf = append(buf, m.Password...)
	}
	return buf, nil
}

// 编码
func (m *RequestMessage) unmarshal(r io.Reader) error {
	var buf [255]byte
	if _, err := io.ReadFull(r, buf[:1]); err != nil {
		return err
	}
	b1 := buf[0]
	// network
	if b1&(1<<4) > 0 {
		m.Network = "udp"
	} else {
		m.Network = "tcp"
	}
	// host length
	hl := 4
	// ipv6
	if b1&(1<<5) > 0 {
		hl = 16
	}
	ip := true
	// hostname
	if b1&(1<<6) > 0 {
		if _, err := io.ReadFull(r, buf[:1]); err != nil {
			return err
		}
		ip = false
		hl = int(buf[0])
	}
	if _, err := io.ReadFull(r, buf[:hl]); err != nil {
		return err
	}
	host := string(buf[:hl])
	if ip {
		host = net.IP(host).String()
	}

	if _, err := io.ReadFull(r, buf[:2]); err != nil {
		return err
	}
	port := binary.BigEndian.Uint16(buf[:2])
	// address
	m.Address = net.JoinHostPort(host, strconv.Itoa(int(port)))
	// Auth
	if b1&1 > 0 {
		if _, err := io.ReadFull(r, buf[:2]); err != nil {
			return err
		}
		ul := buf[0]
		pl := buf[1]
		if ul > 0 {
			if _, err := io.ReadFull(r, buf[:ul]); err != nil {
				return err
			} else {
				m.Username = string(buf[:ul])
			}
		}
		if pl > 0 {
			if _, err := io.ReadFull(r, buf[:pl]); err != nil {
				return err
			} else {
				m.Password = string(buf[:pl])
			}
		}
	}
	return nil
}

type ResponseMessage struct {
	err error
}

func (m *ResponseMessage) Error() error {
	return m.err
}

// 编码
func (m *ResponseMessage) marshal() ([]byte, error) {
	if m.err == nil {
		return []byte{0}, nil
	}
	msg := m.err.Error()
	buf := make([]byte, 0, 1+len(msg))
	buf = append(buf, byte(len(msg)))
	buf = append(buf, msg...)
	return buf, nil
}

// 编码
func (m *ResponseMessage) unmarshal(r io.Reader) error {
	var buf [255]byte
	if _, err := io.ReadFull(r, buf[:1]); err != nil {
		return err
	}
	l := buf[0]
	if l > 0 {
		if _, err := io.ReadFull(r, buf[:l]); err != nil {
			return err
		} else {
			m.err = errors.New(string(buf[:l]))
		}
	}
	return nil
}
