package models

import (
	"log"

	"github.com/gorilla/websocket"
)

type Conn struct {
	ws        *websocket.Conn
	sarteam   *SARTeam
	messageID uint8
}

func (c *Conn) Send(msg string) error {
	if c.ws == nil {
		return nil
	}
	return c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (c *Conn) Start() {
	log.Printf("Starting conn")

	for {
		_, raw, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %s", err)
			return
		}

		log.Printf("%s", raw)

		mutation := mutationFromString(string(raw), c)
		log.Printf("%+v", mutation)
		c.sarteam.mutations <- mutation
	}
}
