package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/rasky/realcrypto/common"

	"github.com/nlopes/slack"
	"github.com/vmihailenco/msgpack"
)

// interactionHandler handles interactive message response.
type interactionHandler struct {
	slackClient       *slack.Client
	imgcache          *ImageCache
	verificationToken string
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	action := message.Actions[0]
	log.Printf("INTERACTION ACTION: %#v", action)
	switch action.Name {
	case actionStart:

		fax := common.Fax{
			Sender:    message.OriginalMessage.Attachments[0].AuthorName,
			Timestamp: time.Now(),
			Message:   action.Value,
		}

		// See if there was a picture
		imageurl := message.OriginalMessage.Attachments[0].ImageURL
		if imageurl != "" {
			if resp, err := http.Get(imageurl); err == nil {
				fax.Picture, _ = ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}

		// Try to send a message to CloudMQTT
		mqtt, err := common.NewMqttClient("backend", env.MqttUrl)
		if err != nil {
			log.Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer mqtt.Disconnect(0)
		payload, err := msgpack.Marshal(&fax)
		if err != nil {
			panic(err) // programming error, structure not marshalable
		}

		token := mqtt.Publish(common.FaxMqttTopic, 2, false, payload)
		if !token.WaitTimeout(5 * time.Second) {
			log.Printf("[ERROR] timeout while publishing to cloudmqtt")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		h.imgcache.Del("/channel/" + message.Channel.ID) // use images once only
		title := ":ok: your fax has been encrypted and transmitted!"
		responseMessage(w, message.OriginalMessage, title, "")
		return
	case actionCancel:
		h.imgcache.Del("/channel/" + message.Channel.ID) // use images once only
		title := fmt.Sprintf(":x: request canceled")
		responseMessage(w, message.OriginalMessage, title, "")
		return
	default:
		log.Printf("[ERROR] Invalid action was submitted: %s", action.Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func responseMessage(w http.ResponseWriter, original slack.Message, title, value string) {
	original.Attachments[0].Actions = []slack.AttachmentAction{} // empty buttons
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}
