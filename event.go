package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type EventHandler func(event Event, c *Client) error

const (
	EventAddFriend   = "add_friend"
	EventNewFriend   = "new_friend"
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
)

type AddFriendEvent struct {
	Channel string `json:"channel"`
}

type NewFriendEvent struct {
	SenderId string `json:"senderId"`
}

type SendMessageEvent struct {
	SenderId string `json:"senderId"`
	Message  string `json:"message"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

// func SendMessageHandler(event Event, c *Client) error {
// 	// Marshal Payload into wanted format
// 	var chatevent SendMessageEvent
// 	if err := json.Unmarshal(event.Data, &chatevent); err != nil {
// 		return fmt.Errorf("bad payload in request: %v", err)
// 	}

// 	// Prepare an Outgoing Message to others
// 	var broadMessage NewMessageEvent

// 	broadMessage.Sent = time.Now()
// 	broadMessage.SenderId = chatevent.SenderId
// 	broadMessage.Message = chatevent.Message

// 	data, err := json.Marshal(broadMessage)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal broadcast message: %v", err)
// 	}

// 	// Place payload into an Event
// 	var outgoingEvent Event
// 	outgoingEvent.Type = EventNewMessage
// 	outgoingEvent.Data = data
// 	// Broadcast to all other Clients
// 	for client := range c.manager.clients {

// 		client.egress <- outgoingEvent

// 	}
// 	return nil
// }

func AddFriendHandler(event Event, c *Client) error {
	var friendEvent AddFriendEvent
	if err := json.Unmarshal(event.Data, &friendEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	var newFriend NewFriendEvent

	newFriend.SenderId = c.userId

	data, err := json.Marshal(newFriend)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	var outgoingEvent Event
	outgoingEvent.Type = EventNewFriend
	outgoingEvent.Data = data

	for client := range c.manager.channels[friendEvent.Channel] {
		client.egress <- outgoingEvent
	}

	return nil
}

func SendMessageHandler(event Event, c *Client) error {
	log.Println(event)

	return nil
}
