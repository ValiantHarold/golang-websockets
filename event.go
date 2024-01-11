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
	EventIncomingFriend = "incoming_friend"
	EventNewFriend      = "new_friend"
	EventAcceptFriend   = "accept_friend"
	EventAddFriend      = "add_friend"
	EventSendMessage    = "send_message"
	EventNewMessage     = "new_message"
)

type IncomingFriendEvent struct {
	UserId string `json:"userId"`
}

type NewFriendEvent struct {
	SenderId string `json:"senderId"`
}
type AcceptFriendEvent struct {
	UserId string `json:"userId"`
}
type AddFriendEvent struct {
	SenderId string `json:"senderId"`
	FriendId string `json:"friendId"`
}

type SendMessageEvent struct {
	UserId  string `json:"userId"`
	Message string `json:"message"`
}

type NewMessageEvent struct {
	SenderId string `json:"senderId"`
	Message  string `json:"message"`
}

func IncomingFriendHandler(event Event, c *Client) error {
	var friendEvent IncomingFriendEvent
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

func AcceptFriendHandler(event Event, c *Client) error {
	var friendEvent AcceptFriendEvent
	if err := json.Unmarshal(event.Data, &friendEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	friend, exists := c.manager.clients[friendEvent.UserId]

	if !exists {
		return fmt.Errorf("Friend does not exists")
	}

	// Send to user
	var AddFriend AddFriendEvent
	AddFriend.SenderId = c.userId
	AddFriend.FriendId = friend.userId

	data, err := json.Marshal(AddFriend)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	var outgoingEvent Event
	outgoingEvent.Type = EventNewFriend
	outgoingEvent.Data = data

	c.egress <- outgoingEvent
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
