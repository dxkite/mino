package transport

import (
	"log"
	"net"
	"strings"
)

// 请求本机地址
func IsLoopbackAddr(addr string) bool {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return IsLoopback(host)
	}
	return false
}

var localIpAddr map[string]struct{}

func init() {
	localIpAddr = map[string]struct{}{}
	if its, _ := net.Interfaces(); its != nil {
		for _, it := range its {
			if ads, err := it.Addrs(); err == nil {
				for _, addr := range ads {
					addrMask := addr.String()
					i := strings.Index(addrMask, "/")
					if v := net.ParseIP(addrMask[:i]); v != nil {
						localIpAddr[string(v)] = struct{}{}
					}
				}
			}
		}
	}
}

// 环回地址
func IsLoopback(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return true
		}
		if _, ok := localIpAddr[string(ip)]; ok {
			return true
		}
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
