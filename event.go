package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type EventHandler func(event Event, c *Client) error

const (
	EventAddFriend    = "add_friend"
	EventAcceptFriend = "accept_friend"
	EventSendMessage  = "send_message"
	EventNewMessage   = "new_message"
)

type AddFriendEvent struct {
	SenderId    string `json:"senderId"`
	SenderEmail string `json:"senderEmail"`
}

type SendMessageEvent struct {
	SenderId string `json:"senderId"`
	Message  string `json:"message"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

func SendMessageHandler(event Event, c *Client) error {
	// Marshal Payload into wanted format
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Data, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Prepare an Outgoing Message to others
	var broadMessage NewMessageEvent

	broadMessage.Sent = time.Now()
	broadMessage.SenderId = chatevent.SenderId
	broadMessage.Message = chatevent.Message

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Place payload into an Event
	var outgoingEvent Event
	outgoingEvent.Type = EventNewMessage
	outgoingEvent.Data = data
	// Broadcast to all other Clients
	for client := range c.manager.clients {

		client.egress <- outgoingEvent

	}
	return nil
}
