package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rasky/realcrypto/common"
	"github.com/vmihailenco/msgpack"
)

const (
	ClientId      = "raspberry"
	ClientMqttQos = 2 // Use MQTT QOS=2 to make sure each message is delivered once
)

func main() {
	surl := os.Getenv("CLOUDMQTT_URL")
	if surl == "" {
		log.Fatal("CLOUDMQTT_URL not defined")
	}

	c, err := common.NewMqttClient(ClientId, surl)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Disconnect(0)

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
	})

	select {} // wait forever
}
