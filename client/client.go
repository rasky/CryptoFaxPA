package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rasky/realcrypto/common"
	rpio "github.com/stianeikeland/go-rpio"
	"github.com/vmihailenco/msgpack"
)

const (
	ClientId      = "client"
	ClientMqttQos = 2 // Use MQTT QOS=2 to make sure each message is delivered once
)

func main() {
	surl := os.Getenv("CLOUDMQTT_URL")
	if surl == "" {
		log.Fatal("CLOUDMQTT_URL not defined")
	}

	// Start polling timezone in background
	go common.PollTimezone()

	c, err := common.NewMqttClient(ClientId, surl)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Disconnect(0)

	chfax := make(chan common.Fax, 16)

	c.Subscribe(common.FaxMqttTopic, ClientMqttQos, func(client mqtt.Client, msg mqtt.Message) {
		var fax common.Fax
		if err := msgpack.Unmarshal(msg.Payload(), &fax); err != nil {
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

		chfax <- fax
	})

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	help_pin := NewRPButton(18)
	blockchain_pin := NewRPButton(17)

	// Main loop: serialize all printing to avoid printing from different
	// goroutines at the same time.
	for {
		select {
		case t := <-help_pin.Edges:
			// Don't process buttons if they were pressed too long ago
			if time.Since(t) < time.Second/2 {
				print_help()
			}
		case t := <-blockchain_pin.Edges:
			// Don't process buttons if they were pressed too long ago
			if time.Since(t) < time.Second/2 {
				print_blockchain()
			}
		case fax := <-chfax:
			// Se non è notte fonda, suona la musichetta del modem mentre
			// inizia a stampare il fax
			t := common.NowHere()
			if t.Hour() >= 9 && t.Hour() <= 20 {
				go exec.Command("play", "modem.ogg").Run()

				// Fai suonare un po' la musichetta prima di iniziare a stampare
				time.Sleep(5 * time.Second)
			}

			print_fax(fax)
		}
	}
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

	common.PrintBytes(buf.Bytes(), true)
}

func print_blockchain() {
	var buf bytes.Buffer

	buf.WriteString("\x1b!\x30") // double-height, double-width
	buf.WriteString("HAHAHAHAHAHAHAHAHAHAHAHA\n")

	common.PrintBytes(buf.Bytes(), true)
}
