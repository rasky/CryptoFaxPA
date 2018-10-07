package main

import (
	"net"
	"sort"
	"strings"
	"time"

	"pifke.org/wpasupplicant"
)

type InterfaceDesc struct {
	Name    string
	Status  string
	IP      string
	Comment string
}

func interfaceByName(name string, idesc string) InterfaceDesc {
	desc := InterfaceDesc{Name: idesc, Status: "DISCONNECTED"}

	intf, err := net.InterfaceByName(name)
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

	var raddrdest string
	if strings.ContainsAny(laddrip, ":") {
		raddrdest = "[2606:4700:4700::1111]:80"
	} else {
		raddrdest = "1.1.1.1:80"
	}

	d := net.Dialer{
		Timeout:   time.Duration(2 * time.Second),
		LocalAddr: laddr,
	}
	if conn, err := d.Dial("tcp", raddrdest); err != nil {
		desc.Status = "NOINTERNET"
		return desc
	} else {
		conn.Close()
		desc.Status = "CONNECTED"
		return desc
	}
}

func WpaCurrentNetwork() (string, error) {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return "", err
	}
	defer w.Close()

	if status, err := w.Status(); err != nil {
		return "", err
	} else {
		return status.SSID(), nil
	}
}

// Scan for Wifi networks -- note that this function is blocking and takes
// a while.
func WpaScan() ([]string, error) {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return nil, err
	}
	defer w.Close()

	if err := w.Scan(); err != nil {
		return nil, err
	}

	// Wait for the end of scanning, or 5 seconds as timeout
	t0 := time.Now()
WaitLoop:
	for time.Since(t0) < 5*time.Second {
		select {
		case evt := <-w.EventQueue():
			if evt.Event == "SCAN_RESULTS" {
				break WaitLoop
			}
		case <-time.After(1 * time.Second):
		}
	}

	res, errs := w.ScanResults()
	if len(errs) != 0 {
		return nil, errs[0]
	}

	seen := make(map[string]bool)
	var ssids []string
	for _, bss := range res {
		s := bss.SSID()
		if !seen[s] {
			ssids = append(ssids, s)
			seen[s] = true
		}
	}

	sort.Strings(ssids)
	return ssids, nil
}

// Returns the list of known networks
func WpaKnownNetworks() ([]string, error) {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return nil, err
	}
	defer w.Close()

	res, err := w.ListNetworks()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var ssids []string
	for _, bss := range res {
		s := bss.SSID()
		if !seen[s] {
			ssids = append(ssids, s)
			seen[s] = true
		}
	}

	sort.Strings(ssids)
	return ssids, nil
}
