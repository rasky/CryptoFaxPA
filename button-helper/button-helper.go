package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/rasky/CryptoFaxPA/common"
	rpio "github.com/stianeikeland/go-rpio"
)

func main() {
	// initialize the GPIO library
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	// connect to the button pins and initialize them
	help_pin := rpio.Pin(22)
	help_pin.Input()
	help_pin.PullOff()
	blockchain_pin := rpio.Pin(23)
	blockchain_pin.Input()
	blockchain_pin.PullOff()

	// loop forever
	for {
		fmt.Println("looping...")
		if help_pin.Read() == rpio.Low {
			fmt.Println("help button pressed")
			print_help()
		}
		if blockchain_pin.Read() == rpio.Low {
			fmt.Println("blockchain button pressed")
			print_blockchain()
		}
		time.Sleep(time.Second / 2)
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
