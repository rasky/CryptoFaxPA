package common

import (
	"fmt"
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}

func NewMqttClient(clientId string, uris string) (mqtt.Client, error) {
	uri, err := url.Parse(uris)
	if err != nil {
		return nil, err
	}

	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(3 * time.Second) {
		return nil, fmt.Errorf("timeout while connecting to CloudMQTT")
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("error connecting to CloudMQTT: %v", err)
	}
	return client, nil
}
