package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	writeWait = 10 * time.Second

	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10

	maxMessageSize int64 = 512
)

type ClientList map[*Client]bool

type Client struct {
	userId  string
	conn    *websocket.Conn
	manager *Manager
	egress  chan Event
}

func NewClient(userId string, conn *websocket.Conn, manager *Manager) *Client {
	log.Println("New Client: ", userId)
	return &Client{
		userId,
		conn,
		manager,
		make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.manager.removeChannel("user__" + c.userId + "__friends")
		c.manager.removeChannel("user__" + c.userId + "__messages")
		c.manager.removeClient(c)
	}()

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling event: %v", err)
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("error handling message: ", err)
		}

	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.leaveChannel(c, "user__"+c.userId+"__friends")
		c.manager.leaveChannel(c, "user__"+c.userId+"__messages")
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingPeriod)

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("conn closed: ", err)

				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("failed to send message: ", err)
			}
			log.Println("sent message: ", message)
		case <-ticker.C:
			log.Println("ping")

			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Println("writemsg err: ", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
