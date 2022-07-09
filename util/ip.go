package util

import (
	"crypto/md5"
	"dxkite.cn/log"
	"encoding/hex"
	"net"
	"sort"
	"strings"
)

// 判断是否请求本机Web服务器
func IsRequestSelf(listen, addr string) bool {
	_, at, _ := net.SplitHostPort(listen)
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if IsLoopback(host) && port == at {
			return true
		}
	}
	return false
}

func IsLocalAddr(addr string) bool {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return IsLoopback(host)
	}
	return false
}

var localIpAddr = map[string]bool{}
var machineId = ""
var hardwareAddr = []net.HardwareAddr{}

func init() {
	machineMacs := []string{}
	if its, _ := net.Interfaces(); its != nil {
		for _, it := range its {
			machineMacs = append(machineMacs, string(it.HardwareAddr))
			if it.Flags&net.FlagLoopback == 0 {
				hardwareAddr = append(hardwareAddr, it.HardwareAddr)
			}
			if ads, err := it.Addrs(); err == nil {
				log.Debug(ads)
				for _, addr := range ads {
					addrMask := addr.String()
					i := strings.Index(addrMask, "/")
					if v := net.ParseIP(addrMask[:i]); v != nil {
						localIpAddr[string(v)] = true
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

func IsLoopbackIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	if _, ok := localIpAddr[string(ip)]; ok {
		return true
	}
	return false
}

// 环回地址
func IsLoopback(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		return IsLoopbackIP(ip)
	}

	if ips, err := net.LookupIP(host); err != nil {
		log.Println("LookupIP", err)
	} else {
		log.Debug("lookup", host, ips)
		for _, ip := range ips {
			if IsLoopbackIP(ip) {
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
