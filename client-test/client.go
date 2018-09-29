package main

import (
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rasky/realcrypto/common"
	"github.com/vmihailenco/msgpack"
)

const (
	ClientId      = "raspberry"
	ClientMqttQos = 1 // Use MQTT QOS=1 to be more resilient on disconnections
)

type Fax struct {
	Sender  string
	Message string
	Picture []byte
}

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
		var fax Fax
		if err := msgpack.Unmarshal(msg.Payload(), &fax); err != nil {
			log.Printf("[ERROR] error decoding mqtt payload: %v", err)
			return
		}

		fmt.Printf("* New 📠 incoming: %+v\n", fax)
	})

	select {} // wait forever
}
