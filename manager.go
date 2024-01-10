package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     checkOrigin,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func checkOrigin(r *http.Request) bool {
	return true

	// origin := r.Header.Get("Origin")

	// switch origin {
	// case "localhost:3000":
	// 	return true
	// default:
	// 	return false
	// }
}

type Manager struct {
	clients  ClientList
	channels map[string]ClientList
	handlers map[string]EventHandler
	sync.RWMutex
}

func NewManager() *Manager {
	log.Println("New Manager")
	m := &Manager{
		clients:  make(ClientList),
		channels: make(map[string]ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
	m.handlers[EventAddFriend] = AddFriendHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("There is no such event type")
	}
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New Connection")

	vars := mux.Vars(r)
	userId := vars["userId"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(userId, conn, m)

	m.addClient(client)
	m.joinChannel(client, "user__"+userId+"__friends")
	m.joinChannel(client, "user__"+userId+"__messages")

	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.conn.Close()
		delete(m.clients, client)
	}
}

func (m *Manager) addChannel(channelName string) {
	m.Lock()
	defer m.Unlock()

	m.channels[channelName] = make(ClientList)
}

func (m *Manager) joinChannel(client *Client, channelName string) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.channels[channelName]; !ok {
		m.channels[channelName] = make(ClientList)
	}

	m.channels[channelName][client] = true
}

func (m *Manager) leaveChannel(client *Client, channelName string) {
	m.Lock()
	defer m.Unlock()

	if channel, ok := m.channels[channelName]; ok {
		if _, ok := channel[client]; ok {
			delete(channel, client)
		}
	}
}

func (m *Manager) removeChannel(channelName string) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.channels[channelName]; ok {
		delete(m.channels, channelName)
	}
	log.Println("Removed Channel: ", channelName)

}
