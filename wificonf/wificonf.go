package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/rasky/CryptoFaxPA/common"
)

var flagListenAddr = flag.String("listen", "127.0.0.1:8080", "address to listen to")

var templ *template.Template

func init() {
	box := packr.NewBox("./assets/templates")

	templ = template.New("Templates")
	templ = template.Must(templ.New("head.html").Parse(box.String("head.html")))
	templ = template.Must(templ.New("footer.html").Parse(box.String("footer.html")))
	templ = template.Must(templ.New("connection.html").Parse(box.String("connection.html")))
	templ = template.Must(templ.New("logo.html").Parse(box.String("logo.html")))
	templ = template.Must(templ.New("home.html").Parse(box.String("home.html")))
	templ = template.Must(templ.New("blockchain.html").Parse(box.String("blockchain.html")))
}

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

type Messages struct {
	m   sync.Mutex
	msg []string
}

func (msg *Messages) Add(m string) {
	msg.m.Lock()
	msg.msg = append(msg.msg, m)
	msg.m.Unlock()
}

func (msg *Messages) Get() []string {
	msg.m.Lock()
	m := msg.msg
	msg.msg = nil
	msg.m.Unlock()
	return m
}

var gMessages Messages

func pageHome(rw http.ResponseWriter, req *http.Request) {
	data := struct {
		Active string
	}{
		"home",
	}

	if err := templ.ExecuteTemplate(rw, "home.html", data); err != nil {
		panic(err)
	}
}

func pageBlockchain(rw http.ResponseWriter, req *http.Request) {
	data := struct {
		Active    string
		NerdInfos []common.BlockchainNerdInfo
		Graph     string
	}{
		"blockchain",
		common.GetBlockchainNerdInfos(),
		common.GetBitcoinAsciiGraph(100, 30),
	}

	if err := templ.ExecuteTemplate(rw, "blockchain.html", data); err != nil {
		panic(err)
	}
}

func pageConnection(rw http.ResponseWriter, req *http.Request) {
	// Trigger a wifi refresh every time the page is opened
	go gScanner.Refresh()

	curwifi, _ := WpaCurrentNetwork()
	known, _ := WpaKnownNetworks()

	data := struct {
		Active      string
		Messages    []string
		Interfaces  []InterfaceDesc
		WifiScan    []string
		WifiKnown   []string
		WifiCurrent string
	}{
		"connection",
		gMessages.Get(),
		[]InterfaceDesc{
			interfaceByName("wlan0", "Wi-Fi"),
			interfaceByName("eth0", "Ethernet"),
			interfaceByName("umts", "GSM / UMTS"),
		},
		gScanner.Networks(),
		known,
		curwifi,
	}

	if curwifi != "" {
		data.Interfaces[0].Comment = "(" + curwifi + ")"
	}

	if err := templ.ExecuteTemplate(rw, "connection.html", data); err != nil {
		panic(err)
	}
}

func pageConnectionAdd(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	name, pass := req.FormValue("name"), req.FormValue("password")
	if err := WpaAddNetwork(name, pass); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(fmt.Sprint(err)))
		return
	}

	gMessages.Add(fmt.Sprintf("The Wi-Fi network %q was successfully added. If it's the strongest, CryptoFaxPA will connect soon.", name))
	http.Redirect(rw, req, "/connection", http.StatusSeeOther)

	// Reset the network interface
	go func() {
		time.Sleep(2 * time.Second)
		exec.Command("sudo", "ifdown", "wlan0").Run()
		time.Sleep(2 * time.Second)
		exec.Command("sudo", "ifup", "wlan0").Run()
	}()
}

func pageConnectionRemove(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := req.FormValue("name")
	if err := WpaRemoveNetwork(name); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(fmt.Sprint(err)))
		return
	}

	gMessages.Add(fmt.Sprintf("The Wi-Fi network %q was successfully removed.", name))
	http.Redirect(rw, req, "/connection", http.StatusSeeOther)

	// Reset the network interface
	go func() {
		time.Sleep(2 * time.Second)
		exec.Command("sudo", "ifdown", "wlan0").Run()
		time.Sleep(2 * time.Second)
		exec.Command("sudo", "ifup", "wlan0").Run()
	}()
}

func main() {
	flag.Parse()
	go gScanner.Run()

	static := packr.NewBox("./assets/html")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(static)))
	http.HandleFunc("/", pageHome)
	http.HandleFunc("/connection", pageConnection)
	http.HandleFunc("/connection/add", pageConnectionAdd)
	http.HandleFunc("/connection/remove", pageConnectionRemove)
	http.HandleFunc("/blockchain", pageBlockchain)

	log.Printf("Listening on %v", *flagListenAddr)
	http.ListenAndServe(*flagListenAddr, nil)

	select {}
}
