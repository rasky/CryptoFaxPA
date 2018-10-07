package main

import (
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var flagListenAddr = flag.String("listen", "127.0.0.1:8080", "address to listen to")

var templ = template.Must(template.New("Templates").ParseGlob("assets/templates/*.html")).Funcs(template.FuncMap{
	"xxxeq": func(a, b interface{}) bool {
		return a == b
	},
})

type InterfaceDesc struct {
	Name   string
	Status string
	IP     string
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

func pageConnection(rw http.ResponseWriter, req *http.Request) {

	data := struct {
		Active     string
		Interfaces []InterfaceDesc
	}{
		"connection",
		[]InterfaceDesc{
			interfaceByName("wlan0", "Wi-Fi"),
			interfaceByName("eth0", "Ethernet"),
			interfaceByName("umts", "GSM / UMTS"),
		},
	}

	if err := templ.ExecuteTemplate(rw, "connection.html", data); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets/html"))))
	http.HandleFunc("/connection", pageConnection)

	log.Printf("Listening on %v", *flagListenAddr)
	http.ListenAndServe(*flagListenAddr, nil)

	select {}
}
