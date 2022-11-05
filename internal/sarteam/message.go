package sarteam

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/websocket"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
)

type Message struct {
	messageID uint8
	command   []string
	ws        *websocket.Conn
}

func (m *Message) Okay() {
	_, err := fmt.Fprintf(m.ws, "%d okay", m.messageID)
	return err
}

func (m *Message) Error(err error) {
	fmt.Fprintf(m.ws, "%d error %s", m.messageID, err.Error())
}

func parseMessage(message string, ws *websocket.Conn) (*Message, error) {
	command := strings.Split(message, " ")
	if len(command) < 2 {
		return nil, ErrInvalidMessage
	}

	id, err := strconv.ParseUint(command[0], 10, 8)
	if err != nil {
		return nil, ErrInvalidMessage
	}

	return &Message{
		messageID: uint8(id),
		command:   command[1:],
		ws:        ws,
	}, nil
}
