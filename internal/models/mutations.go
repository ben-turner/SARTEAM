package models

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidCommand is returned when the command is invalid.
	ErrInvalidCommand = errors.New("invalid command")
)

// mutation represents a single change to the model.
type mutation struct {
	// mutationID is a client-provided ID for the mutation. It is mainly used for
	// correlating responses.
	mutationID uint8
	timestamp  time.Time
	command    []string
	requester  *Conn
}

func (m *mutation) Err(err error) {
	fmt.Fprintf(m.requester.ws, "%d error %s", m.mutationID, err.Error())
}

// Timestamp returns the time the mutation was created.
func (m *mutation) Timestamp() time.Time {
	return m.timestamp
}

// Command returns the command that was used to create the mutation.
func (m *mutation) Command() string {
	return strings.Join(m.command, " ")
}

// newMutation creates a new mutation from the given command.
func newMutation(command string, conn *Conn) *mutation {
	splitCommand := strings.Split(command, " ")

	return &mutation{
		timestamp: time.Now(),
		command:   splitCommand,
		requester: conn,
	}
}
