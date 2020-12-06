package socks5

import (
	"errors"
	"strconv"
)

const (
	Version5                = 0x05
	UsernamePasswordVersion = 0x01
	AuthStatusSucceeded     = 0x00

	AddrTypeIPv4 = 0x01
	AddrTypeFQDN = 0x03
	AddrTypeIPv6 = 0x04

	connectMethod      = 0x01
	bindMethod         = 0x02
	udpAssociateMethod = 0x03

	NoAuthRequiredMethod   = 0x00
	UsernamePasswordMethod = 0x02
)

const (
	succeeded uint8 = iota
	serverFailure
	notAllowed
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

type socks5Err struct {
	code    uint8
	message string
}

var (
	errProtocolVersion = errors.New("error protocol version")
	errProtocolMethods = errors.New("error protocol methods")
	//errServerFailure = socks5Err{serverFailure, "server failure"}
	errNetworkUnreachable   = socks5Err{networkUnreachable, "network unreachable"}
	errHostUnreachable      = socks5Err{hostUnreachable, "host unreachable"}
	errConnectionRefused    = socks5Err{connectionRefused, "connection refused"}
	errCommandNotSupported  = socks5Err{commandNotSupported, "command not supported"}
	errAddrTypeNotSupported = socks5Err{addrTypeNotSupported, "unknown address type"}
)

func (c socks5Err) Code() uint8 {
	return c.code
}

func (c socks5Err) Error() string {
	return c.message
}

type Reply int

func (code Reply) String() string {
	switch code {
	case 0x00:
		return "succeeded"
	case 0x01:
		return "general SOCKS server failure"
	case 0x02:
		return "connection not allowed by ruleset"
	case 0x03:
		return "network unreachable"
	case 0x04:
		return "host unreachable"
	case 0x05:
		return "connection refused"
	case 0x06:
		return "TTL expired"
	case 0x07:
		return "command not supported"
	case 0x08:
		return "address type not supported"
	default:
		return "unknown code: " + strconv.Itoa(int(code))
	}
}
