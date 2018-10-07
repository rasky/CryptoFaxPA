package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

var flagListenAddr = flag.String("listen", "127.0.0.1:8080", "address to listen to")

var templ = template.Must(template.New("Templates").ParseGlob("assets/templates/*.html")).Funcs(template.FuncMap{
	"xxxeq": func(a, b interface{}) bool {
		return a == b
	},
})

type BackgroundScanner struct {
	w     sync.Mutex
	m     sync.Mutex
	ssids []string
}

func (scan *BackgroundScanner) Refresh() {
	scan.w.Lock()
	defer scan.w.Unlock()

	ssids, err := WpaScan()
	if err == nil {
		scan.m.Lock()
		scan.ssids = ssids
		scan.m.Unlock()
	}
}

func (scan *BackgroundScanner) Run() {
	for {
		scan.Refresh()
		time.Sleep(5 * time.Minute)
	}
}

func (scan *BackgroundScanner) Networks() []string {
	scan.m.Lock()
	defer scan.m.Unlock()
	return scan.ssids
}

var gScanner BackgroundScanner

func pageConnection(rw http.ResponseWriter, req *http.Request) {
	// Trigger a wifi refresh every time the page is opened
	go gScanner.Refresh()

	curwifi, _ := WpaCurrentNetwork()
	known, _ := WpaKnownNetworks()

	data := struct {
		Active      string
		Interfaces  []InterfaceDesc
		WifiScan    []string
		WifiKnown   []string
		WifiCurrent string
	}{
		"connection",
		[]InterfaceDesc{
			interfaceByName("wlan0", "Wi-Fi"),
			interfaceByName("eth0", "Ethernet"),
			interfaceByName("umts", "GSM / UMTS"),
		},
		gScanner.Networks(),
		known,
		curwifi,
	}

	data.Interfaces[0].Comment = "(" + curwifi + ")"

	if err := templ.ExecuteTemplate(rw, "connection.html", data); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	go gScanner.Run()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets/html"))))
	http.HandleFunc("/connection", pageConnection)

	log.Printf("Listening on %v", *flagListenAddr)
	http.ListenAndServe(*flagListenAddr, nil)

	select {}
}
