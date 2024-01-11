package main

import (
	"encoding/json"
	"fmt"
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
	UserId string `json:"userId"`
}

type NewFriendEvent struct {
	SenderId string `json:"senderId"`
}

type SendMessageEvent struct {
	UserId  string `json:"userId"`
	Message string `json:"message"`
}

type NewMessageEvent struct {
	SenderId string `json:"senderId"`
	Message  string `json:"message"`
}

func AddFriendHandler(event Event, c *Client) error {
	var friendEvent AddFriendEvent
	if err := json.Unmarshal(event.Data, &friendEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	if _, exists := c.manager.clients[friendEvent.UserId]; !exists {
		return fmt.Errorf("Friend does not exists")
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

	c.manager.clients[friendEvent.UserId].egress <- outgoingEvent

	return nil
}

func SendMessageHandler(event Event, c *Client) error {
	var messageEvent SendMessageEvent
	if err := json.Unmarshal(event.Data, &messageEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	if _, exists := c.manager.clients[messageEvent.UserId]; !exists {
		return fmt.Errorf("Person does not exists")
	}

	var newMessage NewMessageEvent

	newMessage.SenderId = c.userId
	newMessage.Message = messageEvent.Message

	data, err := json.Marshal(newMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	var outgoingEvent Event
	outgoingEvent.Type = EventNewMessage
	outgoingEvent.Data = data

	c.manager.clients[messageEvent.UserId].egress <- outgoingEvent

	return nil
}
