package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
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
	templ = template.Must(templ.New("version.html").Parse(box.String("version.html")))
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
	if req.URL.Path != "/" {
		http.Redirect(rw, req, "/", http.StatusSeeOther)
		return
	}

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
    infos, _ := common.GetBlockchainNerdInfos()
    
	data := struct {
		Active    string
		NerdInfos []common.BlockchainNerdInfo
		Graph     string
	}{
		"blockchain",
		infos,
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
		Interfaces  []common.InterfaceDesc
		WifiScan    []string
		WifiKnown   []string
		WifiCurrent string
	}{
		"connection",
		gMessages.Get(),
		[]common.InterfaceDesc{
			common.InterfaceInspect(common.IntfWiFi),
			common.InterfaceInspect(common.IntfEthernet),
			common.InterfaceInspect(common.IntfGSM),
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

func pageConnectionScan(rw http.ResponseWriter, req *http.Request) {
	go gScanner.Refresh()
	rw.Header().Set("Content-Type", "text/plain")
	io.WriteString(rw, strings.Join(gScanner.Networks(), "\n"))
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

func pageVersion(rw http.ResponseWriter, req *http.Request) {
	var version []byte

	version, _ = ioutil.ReadFile("/var/cache/firmware.last_updated")

	data := struct {
		Active  string
		Version string
	}{
		"version",
		string(version),
	}

	if err := templ.ExecuteTemplate(rw, "version.html", data); err != nil {
		panic(err)
	}
}

func pageVersionUpdate(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("sudo", "/usr/local/sbin/swupdate.sh")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.WriteHeader(200)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		scan := bufio.NewScanner(stdout)
		for scan.Scan() {
			line := scan.Text()
			io.WriteString(rw, fmt.Sprintf("data: %s\n\n", line))
			flusher.Flush()
		}
	}()

	cmd.Wait()
	wg.Wait()
}

func main() {
	flag.Parse()
	go gScanner.Run()

	static := packr.NewBox("./assets/html")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(static)))
	http.HandleFunc("/", pageHome)
	http.HandleFunc("/connection", pageConnection)
	http.HandleFunc("/connection/scan", pageConnectionScan)
	http.HandleFunc("/connection/add", pageConnectionAdd)
	http.HandleFunc("/connection/remove", pageConnectionRemove)
	http.HandleFunc("/blockchain", pageBlockchain)
	http.HandleFunc("/version", pageVersion)
	http.HandleFunc("/version/update", pageVersionUpdate)

	log.Printf("Listening on %v", *flagListenAddr)
	http.ListenAndServe(*flagListenAddr, nil)

	select {}
}
