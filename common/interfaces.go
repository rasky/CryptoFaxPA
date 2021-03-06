package common

import (
	"log"
	"net"
	"strings"
	"syscall"
	"time"
)

type InterfaceDesc struct {
	Name    string
	Status  string
	IP      string
	Comment string
}

type Interface string

const (
	IntfWiFi        Interface = "wlan0"
	IntfAccessPoint Interface = "ap0"
	IntfEthernet    Interface = "eth0"
	IntfGSM         Interface = "eth1"
)

func (i Interface) Name() string {
	switch i {
	case IntfWiFi:
		return "Wi-Fi"
	case IntfAccessPoint:
		return "Access Point"
	case IntfEthernet:
		return "Ethernet"
	case IntfGSM:
		return "GSM / UMTS"
	}
	panic("unreachable")
}

func InterfaceInspect(iif Interface) InterfaceDesc {
	desc := InterfaceDesc{Name: iif.Name(), Status: "DISCONNECTED"}

	intf, err := net.InterfaceByName(string(iif))
	if err != nil || intf.Flags&net.FlagUp == 0 {
		return desc
	}
	ips, err := intf.Addrs()
	if err != nil {
		return desc
	}

	var laddrip string
	var ipstrings []string
	for _, ip := range ips {
		ipinfo, _, _ := net.ParseCIDR(ip.String())
		if ipinfo.IsGlobalUnicast() {
			laddrip = ipinfo.String()
			ipstrings = append(ipstrings, laddrip)
		}
	}
	if len(ipstrings) == 0 {
		return desc
	}

	desc.IP = strings.Join(ipstrings, ", ")

	laddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(laddrip, "0"))
	if err != nil {
		panic(err)
	}

	d := net.Dialer{
		Timeout:   time.Duration(5 * time.Second),
		LocalAddr: laddr,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if err := syscall.BindToDevice(int(fd), string(iif)); err != nil {
					log.Printf("[ERROR] cannot bind to specific interface %v during network testing: %v", string(iif), err)
				}
			})
		},
	}
	if conn, err := d.Dial("tcp", "www.google.com:80"); err != nil {
		desc.Status = "NOINTERNET"
		return desc
	} else {
		conn.Close()
		desc.Status = "CONNECTED"
		return desc
	}
}
