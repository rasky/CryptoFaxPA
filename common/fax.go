package common

import "time"

const FaxMqttTopic = "fax"

type Fax struct {
	Timestamp time.Time
	Sender    string
	Message   string
	Picture   []byte
}
