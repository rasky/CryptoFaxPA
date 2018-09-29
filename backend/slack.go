package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/nlopes/slack"
)

const (
	actionConfirm = "confirm"
	actionStart   = "start"
	actionCancel  = "cancel"
)

type SlackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

// LstenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()

	// Start listening slack events
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	// Only response mention to bot. Ignore else.
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", s.botID)) {
		return nil
	}

	// Parse message
	m := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	if len(m) == 0 || len(m[0]) == 0 {
		return fmt.Errorf("invalid message")
	}

	u, err := s.client.GetUserInfo(ev.Msg.User)
	if err != nil {
		return fmt.Errorf("error retrieving user info: %v", err)
	}

	// value is passed to message handler when request is approved.
	attachment := slack.Attachment{
		Pretext:    "Confirm sending this text to Cryptofax? :fax:",
		AuthorName: u.RealName,
		Color:      "#f9a41b",
		CallbackID: "cryptofax",
		Text:       m[0],
		Actions: []slack.AttachmentAction{
			{
				Name:  actionStart,
				Text:  "Fax it :fax:",
				Type:  "button",
				Value: m[0],
				Style: "primary",
			},
			{
				Name:  actionCancel,
				Text:  "No",
				Type:  "button",
				Style: "danger",
			},
		},
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}

	if _, _, err := s.client.PostMessage(ev.Channel, "", params); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil
}
