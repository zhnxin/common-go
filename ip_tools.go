package common

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

//0101A8C0
func IPv4FromHex(str16ip string) (string, error) {
	b, err := hex.DecodeString(str16ip)
	if err != nil {
		return "", err
	}
	if len(b) != 4 {
		return "", fmt.Errorf("Parse str to ipv4 fail")
	}
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3]), nil
}

func GetInternalIP() (string, error) {
	cnn, err := net.Dial("udp", "10.255.255.255:1")
	if err != nil {
		return "", err
	}
	defer cnn.Close()
	addr := cnn.LocalAddr().String()
	return addr[:strings.LastIndex(addr, ":")], nil
}

func GetAllIP() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ipstr := ip.String()
			if ipstr != "127.0.0.1" && ipstr != "::1" {
				res = append(res, ipstr)
			}
		}
	}

	return res, nil
}
