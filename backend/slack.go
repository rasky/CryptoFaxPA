package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

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

	isIMLock sync.Mutex
	isIM     map[string]bool
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
	// Ignore all kind of special messages that are not real messages, at least
	// for now.
	if ev.Msg.SubType != "" {
		return nil
	}

	// Check if the channel is an IM (using a cache to avoid polling Slack API too much)
	s.isIMLock.Lock()
	if s.isIM == nil {
		s.isIM = make(map[string]bool)
	}
	if _, found := s.isIM[ev.Msg.Channel]; !found {
		channel, err := s.client.GetConversationInfo(ev.Msg.Channel, false)
		if err != nil {
			s.isIMLock.Unlock()
			return fmt.Errorf("error retrieving group info: %#v", err)
		}
		s.isIM[ev.Msg.Channel] = channel.IsIM
	}
	isIm := s.isIM[ev.Msg.Channel]
	s.isIMLock.Unlock()

	m := strings.TrimSpace(ev.Msg.Text)

	// Only response mention to bot. Ignore else (unless in direct message)
	mention := fmt.Sprintf("<@%s> ", s.botID)
	if strings.HasPrefix(m, mention) {
		// Remove mention from text
		m = m[len(mention):]
	} else if !isIm {
		return nil
	}

	// Get information on the user
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
		Text:       m,
		Actions: []slack.AttachmentAction{
			{
				Name:  actionStart,
				Text:  "Fax it :fax:",
				Type:  "button",
				Value: m,
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
