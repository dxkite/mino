package mino

import (
	"errors"
	"net"
)

// 连接信息
type RequestMessage struct {
	// 网络信息
	Network uint8
	// 地址
	Address string
	// 用户名
	Username string
	// 密码
	Password string
	// 硬件地址
	MacAddress []net.HardwareAddr
}

// 编码
func (m *RequestMessage) marshal() []byte {
	buf := []byte{m.Network}
	buf = append(buf, byte(len(m.Address)))
	buf = append(buf, m.Address...)
	buf = append(buf, byte(len(m.Username)))
	buf = append(buf, m.Username...)
	buf = append(buf, byte(len(m.Password)))
	buf = append(buf, m.Password...)
	buf = append(buf, byte(len(m.MacAddress)))
	for _, mac := range m.MacAddress {
		buf = append(buf, mac...)
	}
	return buf
}

// 编码
func (m *RequestMessage) unmarshal(p []byte) (err error) {
	if len(p) < 5 {
		return errors.New("sort message")
	}
	m.Network = p[0]
	off := 1
	if m.Address, off, err = readString(off, p); err != nil {
		return err
	}
	if m.Username, off, err = readString(off, p); err != nil {
		return err
	}
	if m.Password, off, err = readString(off, p); err != nil {
		return err
	}
	m.MacAddress = []net.HardwareAddr{}
	lm := int(p[off])
	off++
	for i := 0; i < lm; i++ {
		if len(p) < off+6 {
			return errors.New("sort message")
		}
		m.MacAddress = append(m.MacAddress, p[off:off+6])
		off += 6
	}
	return nil
}

// 读字符
func readString(off int, p []byte) (string, int, error) {
	l := int(p[off])
	if l == 0 {
		return "", off + 1, nil
	}
	if len(p) < l {
		return "", 0, errors.New("sort message")
	}
	n := off + int(p[off]) + 1
	return string(p[off+1 : n]), n, nil
}

// 响应信息
type ResponseMessage struct {
	Code    uint8
	Message string
}

const (
	succeeded uint8 = iota
	serverFailure
	notAllowed
	networkUnreachable
	hostUnreachable
	connectionRefused
	unknownError
)

type tlsError uint8

func (e tlsError) Error() string {
	switch v := uint8(e); v {
	case serverFailure:
		return "serverFailure"
	case notAllowed:
		return "notAllowed"
	case networkUnreachable:
		return "networkUnreachable"
	case hostUnreachable:
		return "hostUnreachable"
	case connectionRefused:
		return "connectionRefused"
	}
	return "unknown"
}

// 编码
func (m *ResponseMessage) marshal() []byte {
	buf := []byte{m.Code, byte(len(m.Message))}
	buf = append(buf, m.Message...)
	return buf
}

// 编码
func (m *ResponseMessage) unmarshal(p []byte) error {
	if len(p) < 2 {
		return errors.New("sort message")
	}
	m.Code = p[0]
	if str, _, err := readString(1, p); err != nil {
		return err
	} else {
		m.Message = str
	}
	return nil
}
