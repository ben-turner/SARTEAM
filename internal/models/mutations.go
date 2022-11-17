package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidCommand is returned when the command is invalid.
	ErrInvalidCommand = errors.New("invalid command")

	// ErrTeamNotFound is returned when the team is not found.
	ErrTeamNotFound = errors.New("team not found")

	// ErrPermissionDenied is returned when the user does not have permission to perform the action.
	ErrPermissionDenied = errors.New("permission denied")
)

// mutation represents a single change to the model.
type mutation struct {
	// mutationID is a client-provided ID for the mutation. It is mainly used for
	// correlating responses.
	mutationID uint8

	// timestamp is the time the mutation was created.
	timestamp time.Time

	// command is the command that was used to create the mutation.
	command []string

	// undoFuncs is a list of functions that can be used to undo the mutation.
	undoFuncs []func()

	// requester is the connection that requested the mutation.
	requester *Conn
}

func (m *mutation) Error(err error) {
	if m.requester == nil {
		return
	}
	m.requester.Send(fmt.Sprintf("%d error %s", m.mutationID, err.Error()))
}

func (m *mutation) Reply(message string) {
	if m.requester == nil {
		return
	}
	m.requester.Send(fmt.Sprintf("%d %s", m.mutationID, message))
}

// Timestamp returns the time the mutation was created.
func (m *mutation) Timestamp() time.Time {
	return m.timestamp
}

// Command returns the command that was used to create the mutation.
func (m *mutation) Command() string {
	return strings.Join(m.command, " ")
}

// Pops the first n elements from the command and returns them.
func (m *mutation) Pop(n int) []string {
	if len(m.command) < n {
		return nil
	}

	popped := m.command[:n]
	m.command = m.command[n:]
	return popped
}

// Undo undoes the mutation.
func (m *mutation) Undo() {
	for _, undoFunc := range m.undoFuncs {
		undoFunc()
	}
}

// LogMessage returns a string representation of the mutation.
func (m *mutation) LogMessage() string {
	return fmt.Sprintf("%s %s", m.timestamp.Format(time.RFC3339), m.Command())
}

// mutationFromSlice creates a new mutation from the given command.
func mutationFromSlice(command []string, conn *Conn) *mutation {
	timestamp, err := time.Parse(time.RFC3339, command[0])

	if err != nil {
		timestamp = time.Now()
	}

	id, err := strconv.ParseUint(command[0], 10, 8)
	if err == nil {
		command = command[1:]
	}

	return &mutation{
		timestamp:  timestamp,
		command:    command,
		requester:  conn,
		mutationID: uint8(id),
	}
}

// mutationFromString creates a new mutation from the given command.
func mutationFromString(command string, conn *Conn) *mutation {
	splitCommand := strings.Split(command, " ")

	return mutationFromSlice(splitCommand, conn)
}
