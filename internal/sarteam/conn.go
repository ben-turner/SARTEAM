package sarteam

import (
	"bufio"

	"golang.org/x/net/websocket"
)

type Conn struct {
	ws      *websocket.Conn
	sarteam *SARTeam
	messageID uint8
}



func (c *Conn) Start() {
	scanner := bufio.NewScanner(c.ws)
	for scanner.Scan() {
		raw = scanner.Text()
		msg, err := parseMessage(raw, c.ws)
		if err != nil {
			continue
		}

		switch msg.command[0] {

		}

		id := splitMsg[0]

		err := c.sarteam.RootModel.Update(scanner.Text())
		if err != nil {
			c.
	}
}
