package util

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"net"
	"sort"
	"strings"
)

// 判断是否是本机IP
func IsRequestHttp(listen, addr string) bool {
	_, at, _ := net.SplitHostPort(listen)
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if IsLoopback(host) && port == at {
			return true
		}
	}
	return false
}

var localIpAddr map[string]struct{}
var machineId = ""
var hardwareAddr []net.HardwareAddr

func init() {
	localIpAddr = map[string]struct{}{}
	machineMacs := []string{}
	if its, _ := net.Interfaces(); its != nil {
		for _, it := range its {
			machineMacs = append(machineMacs, string(it.HardwareAddr))
			if it.Flags&net.FlagLoopback == 0 {
				hardwareAddr = append(hardwareAddr, it.HardwareAddr)
			}
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

	sort.Strings(machineMacs)
	h := md5.New()
	h.Write([]byte(strings.Join(machineMacs, "")))
	machineId = hex.EncodeToString(h.Sum(nil))
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

// 获取机械码
func GetMachineId() string {
	return machineId
}

// 检查机器白名单
func CheckMachineId(id string) bool {
	if len(id) == 0 {
		return true
	}
	return machineId == id
}

// 获取网卡地址
func GetHardwareAddr() []net.HardwareAddr {
	return hardwareAddr
}
