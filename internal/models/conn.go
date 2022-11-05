package models

import (
	"bufio"

	"golang.org/x/net/websocket"
)

type Conn struct {
	ws        *websocket.Conn
	sarteam   *SARTeam
	messageID uint8
}

func (c *Conn) Send(msg string) error {
	_, err := c.ws.Write([]byte(msg))
	return err
}

func (c *Conn) Start() {
	scanner := bufio.NewScanner(c.ws)
	for scanner.Scan() {
		mutation := newMutation(scanner.Text(), c)
		c.sarteam.ApplyMutation(mutation)
	}
}
