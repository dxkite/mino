package transport

import (
	"log"
	"net"
)

// 请求本机地址
func IsLoopbackAddr(addr string) bool {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return IsLoopback(host)
	}
	return false
}

// 环回地址
func IsLoopback(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	if ips, err := net.LookupIP(host); err != nil {
		log.Println("LookupIP", err)
	} else {
		for _, ip := range ips {
			if ip.IsLoopback() {
				return true
			}
		}
	}
	return false
}
