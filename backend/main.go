package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

// https://api.slack.com/slack-apps
// https://api.slack.com/internal-integrations
type envConfig struct {
	// URL at which the server is reachable
	ServerUrl string `envconfig:"SERVER_URL" required:"true"`

	// Port is server port to be listened.
	Port string `envconfig:"PORT" default:"3000"`

	// BotToken is bot user token to access to slack API.
	BotToken string `envconfig:"BOT_TOKEN" required:"true"`

	// VerificationToken is used to validate interactive messages from slack.
	VerificationToken string `envconfig:"VERIFICATION_TOKEN" required:"true"`

	// BotID is bot user ID.
	BotID string `envconfig:"BOT_ID" required:"true"`

	// CloudMQTT URL to connect to
	MqttUrl string `envconfig:"CLOUDMQTT_URL" required:"true"`

	// Redis URL to connect to
	RedisUrl string `envconfig:"REDIS_URL" required:"true"`

	// Turn off low-level Slack API debugging
	Debug bool `envconfig:"DEBUG"`
}

var env envConfig

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return 1
	}

	imgcache, err := NewImageCache(env.RedisUrl)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to Redis: %s", err)
		return 1
	}

	// Listening slack event and response
	log.Printf("[INFO] Start slack event listening")
	client := slack.New(env.BotToken)
	client.SetDebug(env.Debug)
	slackListener := &SlackListener{
		token:    env.BotToken,
		client:   client,
		botID:    env.BotID,
		imgcache: imgcache,
	}
	go slackListener.ListenAndResponse()

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	http.Handle("/interaction", interactionHandler{
		verificationToken: env.VerificationToken,
		imgcache:          imgcache,
	})

	http.HandleFunc("/image/", func(rw http.ResponseWriter, req *http.Request) {
		var img []byte
		if err := imgcache.Get(req.URL.Path, &img); err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		rw.Header().Set("Content-type", "image/png")
		io.Copy(rw, bytes.NewReader(img))
	})

	log.Printf("[INFO] Server listening on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		log.Printf("[ERROR] %s", err)
		return 1
	}

	return 0
}
