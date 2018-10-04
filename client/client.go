package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	humanize "github.com/dustin/go-humanize"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rasky/CryptoFaxPA/common"
	"github.com/vmihailenco/msgpack"
)

const (
	ClientId      = "client"
	ClientMqttQos = 2 // Use MQTT QOS=2 to make sure each message is delivered once

	PinHelp       = 22
	PinBlockchain = 23
)

var (
	flagSpoolDir = flag.String("spool", "/var/spool/cryptofax", "spool directory to use")
)

func main() {
	flag.Parse()

	surl := os.Getenv("CLOUDMQTT_URL")
	if surl == "" {
		log.Fatal("CLOUDMQTT_URL not defined")
	}

	if fi, err := os.Stat(*flagSpoolDir); err != nil || !fi.IsDir() {
		log.Fatalf("%s does not exist or is not a directory", *flagSpoolDir)
	}

	// Start background bootstrap sound
	go exec.Command("play", "startup.ogg").Run()

	// Start polling timezone in background
	go common.PollTimezone()

	chfax := make(chan bool, 16)

	// See if there are pending faxes in the spool; if so, schedule them right away
	if files, err := ioutil.ReadDir(*flagSpoolDir); err == nil && len(files) > 0 {
		go func() {
			for _ = range files {
				chfax <- true
			}
		}()
	}

	go PollMqtt(chfax, surl)

	buttonMonitor := NewRPButtonMonitor(PinHelp, PinBlockchain)
	defer buttonMonitor.Shutdown()

	// Wait for startup sound to finish before begin processing events.
	// This avoids the modem sound to play over the startup sound in case
	// a fax is immediately available after boot.
	time.Sleep(5 * time.Second)

	// Main loop: serialize all printing to avoid printing from different
	// goroutines at the same time.
	for {
		select {
		case evt := <-buttonMonitor.Events:
			// Don't process buttons if they were pressed too long ago
			if time.Since(evt.When) > time.Second/2 {
				continue
			}
			switch evt.Pin {
			case PinHelp:
				fmt.Println("help button pressed")
				print_help()
			case PinBlockchain:
				fmt.Println("blockchain button pressed")
				print_blockchain()
			}
		case <-chfax:
			print_fax_from_spool()
		}
	}
}

func PollMqtt(chfax chan bool, surl string) {
	var c mqtt.Client
	sleep := 5 * time.Second
	for {
		var err error
		c, err = common.NewMqttClient(ClientId, surl)
		if err != nil {
			log.Printf("[INFO] cannot connect to MQTT server: %v", err)
			log.Printf("[INFO] retrying in %v...", sleep)
			time.Sleep(sleep)
			sleep = sleep + sleep/3
			if sleep > 5*time.Minute {
				sleep = 5 * time.Minute
			}
		} else {
			break
		}
	}
	defer c.Disconnect(0)

	c.Subscribe(common.FaxMqttTopic, ClientMqttQos, func(client mqtt.Client, msg mqtt.Message) {
		// Use a filename whose alphabetical sorting respects the order of arrival
		filename := fmt.Sprintf("%s/%016x", *flagSpoolDir, time.Now().Unix())
		common.WriteFileSync(filename, msg.Payload(), 0777)
		chfax <- true
	})

	log.Printf("[INFO] connected to MQTT server, start polling")
	select {}
}

func print_fax_from_spool() {
	files, err := ioutil.ReadDir(*flagSpoolDir)
	if err != nil {
		log.Printf("[ERROR] cannot access spool dir: %v", err)
		return
	}

	// From this point on, remove the file after we finished processing it.
	// If the process crashes (eg: system shutdown), the file will still be there
	fn := *flagSpoolDir + "/" + files[0].Name()
	defer os.Remove(fn)

	payload, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Printf("[ERROR] cannot read spool file: %v", err)
		return
	}

	var fax common.Fax
	if err := msgpack.Unmarshal(payload, &fax); err != nil {
		log.Printf("[ERROR] error decoding mqtt payload: %v", err)
		return
	}

	fmt.Printf("* New 📠 incoming:\n")
	fmt.Printf("    - Sender: %v\n", fax.Sender)
	fmt.Printf("    - Timestamp: %v\n", fax.Timestamp)
	fmt.Printf("    - Message: %v\n", fax.Message)
	if len(fax.Picture) != 0 {
		fmt.Printf("    - Picture: %v bytes\n", len(fax.Picture))
		if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
			fmt.Println()
			fmt.Printf("\x1b]1337;File=width=40%%;inline=1:%s\x07\n", base64.StdEncoding.EncodeToString(fax.Picture))
		}
	}

	common.StartBlinking()
	defer common.StopBlinking()

	// Se non è notte fonda, suona la musichetta del modem mentre
	// inizia a stampare il fax
	t := common.NowHere()
	if t.Hour() >= 9 && t.Hour() <= 20 {
		go exec.Command("play", "modem.ogg").Run()

		// Fai suonare un po' la musichetta prima di iniziare a stampare
		time.Sleep(6 * time.Second)
	}

	print_fax(fax)
}

func print_fax(fax common.Fax) {
	var buf bytes.Buffer
	buf.WriteString("\x1b!\x10") // double-height
	fmt.Fprintf(&buf, "Fax from ")
	buf.WriteString("\x1b!\x90") // double-height, underlined
	fmt.Fprintf(&buf, "%v\n", fax.Sender)

	buf.WriteString("\x1b!\x00") // font A, single-height
	fmt.Fprintf(&buf, "(%v)\n\n", fax.Timestamp.Format("2006-01-02 15:04"))

	if fax.Message != "" {
		buf.Write(common.EncodeForPrinter(fax.Message))
		buf.WriteString("\n")
	}

	if len(fax.Picture) != 0 {
		common.PrintBytes(buf.Bytes(), false)
		common.PrintImage(fax.Picture, true)
	} else {
		common.PrintBytes(buf.Bytes(), true)
	}
}

func print_help() {
	common.StartBlinking()
	defer common.StopBlinking()

	var buf bytes.Buffer

	buf.WriteString("\x1b!\x30") // double-height, double-width
	buf.WriteString("CryptoFaxPA\n")

	buf.WriteString("\x1b!\x00") // font A, single-height
	buf.Write(common.EncodeForPrinter(`CryptoFaxPA è il nuovo dispositivo pensato per la comunicazione digitale della pubblica amministrazione. Caratterizzato da un design elegante e minimale, è adatto ai moderni uffici di dirigenti e funzionari pubblici ma anche, perché no, alle loro abitazioni, dove sempre più spesso si svolge la comune attività lavorativa.
In particolare CryptoFaxPA consente all'utente (d'ora in avanti denominato per semplicità "Diego") di rimanere aggiornato su tutte le novità e gli umori del suo amato Team.`))
	buf.WriteString("\n\n")

	buf.WriteString("\x1b!\x80") // font A, underlined
	buf.WriteString("Come funziona\n")
	buf.WriteString("\x1b!\x00") // font A, single-height
	buf.Write(common.EncodeForPrinter(`Basterà mantenere il dispositivo connesso alla rete attraverso il WiFi o la rete GSM (inclusa) per ricevere, in qualsiasi momento, i preziosi messaggi.`))
	buf.WriteString("\n\n")

	buf.WriteString("\x1b!\x80") // font A, underlined
	buf.WriteString("Configurazione WiFi\n")
	buf.WriteString("\x1b!\x00") // font A, single-height
	buf.Write(common.EncodeForPrinter(`Se CryptoFaxPA non rileva una rete WiFi nota, trascorsi 120 secondi si avvia in modalità access point esponendo una rete wireless di nome CryptoFaxPA. A quel punto basterà accedervi con un qualsiasi altro device ed aprire la pagina http://cryptofaxpa.local, dove sarà possibile configurare la propria rete WiFi.`))

	// print network addresses
	out, err := exec.Command("/bin/bash", "-c", `/sbin/ifconfig | /usr/bin/awk -v RS="\n\n" '{ for (i=1; i<=NF; i++) if ($i == "inet") address = $(i+1); if (address != "127.0.0.1") printf "%s\t%s\n", $1, address }'`).Output()
	if err == nil {
		buf.WriteString("\n\n")
		buf.Write(out)
		buf.WriteString("\n\n")
	}

	common.PrintBytes(buf.Bytes(), true)
}

func httpGetString(url string) string {
	res, err := http.Get(url)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	io.Copy(&buf, res.Body)
	res.Body.Close()
	return buf.String()
}

func satoshis(s string) string {
	v, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return fmt.Sprintf("%.8f", float64(v)/float64(100000000))
	}
	return s
}

func seconds(s string) string {
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return fmt.Sprintf("%v", time.Duration(v*1000000000))
	}
	return s
}

func human(s string) string {
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return fmt.Sprintf("%v", humanize.Comma(int64(v)))
	}
	return s
}

func print_blockchain() {
	common.StartBlinking()
	defer common.StopBlinking()

	var buf bytes.Buffer

	buf.WriteString("\x1b!\x30") // double-height, double-width
	buf.WriteString("BLOCKCHAIN SUPER NERD INFO\n")
	buf.WriteString("\x1b!\x00") // font A, single-height
	fmt.Fprintln(&buf, "Updated at:", common.NowHere())

	fmt.Fprintln(&buf, "Current BTC price (USD):\n", "$"+httpGetString("https://blockchain.info/q/24hrprice"))
	fmt.Fprintln(&buf, "Market cap (USD):\n", "$"+human(httpGetString("https://blockchain.info/q/marketcap")))
	fmt.Fprintln(&buf, "Global hash rate (GigaHash):\n", human(httpGetString("https://blockchain.info/q/hashrate")))
	fmt.Fprintln(&buf, "Current difficulty target:\n", human(httpGetString("https://blockchain.info/q/getdifficulty")))
	fmt.Fprintln(&buf, "Current block height:\n", httpGetString("https://blockchain.info/q/getblockcount"))
	fmt.Fprintln(&buf, "Latest hash:\n", httpGetString("https://blockchain.info/q/latesthash"))
	fmt.Fprintln(&buf, "Current block reward:\n", satoshis(httpGetString("https://blockchain.info/q/bcperblock")))
	fmt.Fprintln(&buf, "Total bitcoins:\n", satoshis(httpGetString("https://blockchain.info/q/totalbc")))
	fmt.Fprintln(&buf, "Probability of mining:\n", httpGetString("https://blockchain.info/q/probability"))
	fmt.Fprintln(&buf, "ETA until next block:\n", seconds(httpGetString("https://blockchain.info/q/eta")))
	common.PrintBytes(buf.Bytes(), true)

	graph := GetBitcoinGraph()
	if graph != nil {
		var buf bytes.Buffer
		buf.WriteString("\x1b!\x30") // double-height, double-width
		buf.WriteString("BLOCKCHAIN LIVE EXCHANGE\n")
		buf.WriteString("\x1b!\x00") // font A, single-height
		common.PrintBytes(buf.Bytes(), false)

		common.PrintImage(graph, true)
	}
}
