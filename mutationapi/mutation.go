package mutationapi

import (
	"encoding/json"
	"strings"
	"time"
)

type (
	// MutationID is a mostly unique identifier for a mutation. Depending on the
	// source of the mutation, the ID may be a UUID or a simple incrementing ID.
	// IDs are not guaranteed to be unique and a single mutation may change IDs
	// depending on the Conn that it is sent to. The ID is prefixed with a '-' if
	// the mutation is an inverse of another mutation.
	MutationID string

	// MutationAction is the type of mutation that is being performed.
	MutationAction uint8
)

const (
	// MutationActionUnknown is an unknown/invalid mutation action.
	MutationActionUnknown MutationAction = iota
	// MutationActionCreate is a mutation that creates a new resource. The body of
	// the mutation should be the new resource. The path may or may not include an
	// id for the resource. If the mutation path does not include an id, the
	// mutation path should be modified to include the id when the mutation is
	// applied. The inverse of this mutation is a DELETE mutation.
	MutationActionCreate
	// MutationActionRead is a mutation that reads a resource. It's worth noting
	// that this is not technically a "mutation" so much as a remote procedure
	// call. Read mutations must not modify the state.
	MutationActionRead
	// MutationActionUpdate is a mutation that updates a resource. The body of the
	// mutation should be the new value for the specified path. The inverse of
	// this mutation is an UPDATE mutation with the original body.
	MutationActionUpdate
	// MutationActionDelete is a mutation that deletes a resource. The inverse of
	// this mutation is a CREATE mutation with the original body and resource id.
	MutationActionDelete
)

// String returns a string representation of the mutation action.
func (m MutationAction) String() string {
	switch m {
	case MutationActionCreate:
		return "CREATE"
	case MutationActionRead:
		return "READ"
	case MutationActionUpdate:
		return "UPDATE"
	case MutationActionDelete:
		return "DELETE"
	}

	return "UNKNOWN"
}

func ParseMutationAction(action string) MutationAction {
	switch action {
	case "CREATE":
		return MutationActionCreate
	case "READ":
		return MutationActionRead
	case "UPDATE":
		return MutationActionUpdate
	case "DELETE":
		return MutationActionDelete
	}

	return MutationActionUnknown
}

type Mutation struct {
	// ID is a client-provided identifier that can be used to correlate responses.
	ID MutationID

	// Timestamp is the time the mutation was created.
	Timestamp time.Time

	// Conn is the connection that requested the mutation.
	Conn Conn

	// Action is the action that the mutation is performing.
	Action MutationAction

	// Path is the path that the mutation is acting on.
	Path []string

	// OriginalBody stores the original value of a given path prior to the
	// mutation being applied. It is used to revert the mutation if it fails.
	OriginalBody []byte

	// body is an arbitrary value that can be used by the mutation.
	body []byte
}

// BodyAsBytes returns the body data as a byte slice.
func (m *Mutation) BodyAsBytes() []byte {
	return m.body
}

// BodyAsJSON parses the body data as JSON into the provided pointer value.
func (m *Mutation) BodyAsJSON(v any) error {
	return json.Unmarshal(m.BodyAsBytes(), v)
}

// BodyAsString returns the body data as a string.
func (m *Mutation) BodyAsString() string {
	return string(m.BodyAsBytes())
}

// BodyAsBool returns the body as a boolean value.
func (m *Mutation) BodyAsBool() bool {
	return m.BodyAsString() == "true"
}

// Error marks the mutation as failed and sends the error message to the client.
func (m *Mutation) Error(err error) {
	m.Conn.sendError(m.ID, err)
}

// String returns a string representation of the mutation.
func (m *Mutation) String() string {
	return strings.Join([]string{
		m.Timestamp.Format(time.RFC3339),
		string(m.ID),
		m.Action.String(),
		strings.Join(m.Path, "/"),
		m.BodyAsString(),
	}, " ")
}

func invertID(id MutationID) MutationID {
	if id[0] == '-' {
		return id[1:]
	}

	return "-" + id
}

// Inverse returns a new mutation that would undo the current mutation.
//
// Calling this method prior to setting the OriginalBody will result in
// undefined behavior.
//
// This method will return nil if the mutation is not reversible or if the
// original mutation does not modify the state (eg. GET mutations).
func (m *Mutation) Inverse() *Mutation {
	switch m.Action {
	case MutationActionCreate:
		return &Mutation{
			ID:           invertID(m.ID),
			Timestamp:    time.Now(),
			Conn:         m.Conn,
			Action:       MutationActionDelete,
			Path:         m.Path,
			OriginalBody: m.body,
		}
	case MutationActionUpdate:
		return &Mutation{
			ID:           invertID(m.ID),
			Timestamp:    time.Now(),
			Conn:         m.Conn,
			Action:       MutationActionUpdate,
			Path:         m.Path,
			OriginalBody: m.body,
			body:         m.OriginalBody,
		}
	case MutationActionDelete:
		return &Mutation{
			ID:        invertID(m.ID),
			Timestamp: time.Now(),
			Conn:      m.Conn,
			Action:    MutationActionCreate,
			Path:      m.Path,
			body:      m.OriginalBody,
		}
	}

	return nil
}

// ParseMutation parses a mutation from a string.
//
// The string should be in the following format:
// [timestamp] <id> <action> <path> [body]
func ParseMutation(msg string, conn Conn) *Mutation {
	split := strings.Split(msg, " ")
	if len(split) < 2 { // must have at least action and path
		return nil
	}

	timestamp, err := time.Parse(time.RFC3339, split[0])
	if err != nil {
		timestamp = time.Now()
	} else {
		split = split[1:]
	}

	id := MutationID(split[0])

	action := ParseMutationAction(split[0])

	path := strings.Split(split[1], "/")

	body := []byte(strings.Join(split[2:], " "))

	return &Mutation{
		ID:        id,
		Timestamp: timestamp,
		Conn:      conn,
		Action:    action,
		Path:      path,
		body:      body,
	}
}
