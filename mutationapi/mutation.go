package mutationapi

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// MutationID is a mostly unique identifier for a mutation. Depending on the
// source of the mutation, the ID may be a UUID or a simple incrementing ID.
// IDs are not guaranteed to be unique and a single mutation may change IDs
// depending on the Conn that it is sent to. The ID is prefixed with a '-' if
// the mutation is an inverse of another mutation.
type MutationID string

// MutationAction is the type of mutation that is being performed.
type MutationAction uint8

// Path represents a path in the mutation tree.
type Path []string

func (p Path) String() string {
	return strings.Join(p, "/")
}

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
	// ClientID is the ID the client assigned to the mutation. This is used to
	// match responses to requests.
	ClientID MutationID

	// ID is unique identifier for the mutation. The ID is prefixed with a '-'
	// if the mutation is an inverse of another mutation.
	ID MutationID

	// Timestamp is the time the mutation was created.
	Timestamp time.Time

	// Conn is the connection that requested the mutation.
	Conn Conn

	// Action is the action that the mutation is performing.
	Action MutationAction

	// Path is the path that the mutation is acting on.
	Path Path

	// OriginalBody stores the original value of a given path prior to the
	// mutation being applied. It is used to revert the mutation if it fails.
	OriginalBody []byte

	// Body is an arbitrary value that can be used by the mutation.
	Body []byte
}

// BodyAsBytes returns the body data as a byte slice.
func (m *Mutation) BodyAsBytes() []byte {
	return m.Body
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
func (m *Mutation) BodyAsBool() (bool, bool) {
	switch m.BodyAsString() {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}

// Error marks the mutation as failed and sends the error message to the client.
func (m *Mutation) Error(err error) {
	if m == nil {
		return
	}

	m.Conn.sendError(err)
}

// String returns a string representation of the mutation.
func (m *Mutation) String(useClientID bool) string {
	parts := make([]string, 4, 5)
	parts[0] = m.Timestamp.UTC().Format(time.RFC3339)
	parts[1] = string(m.ID)
	if useClientID {
		parts[1] = string(m.ClientID)
	}
	parts[2] = m.Action.String()
	parts[3] = strings.Join(m.Path, "/")

	if len(m.Body) > 0 {
		parts = append(parts, string(m.Body))
	}

	return strings.Join(parts, " ")
}

func invertID(id MutationID) MutationID {
	if len(id) == 0 {
		return id
	}

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
			ClientID:     invertID(m.ClientID),
			ID:           invertID(m.ID),
			Timestamp:    timeNow(),
			Conn:         m.Conn,
			Action:       MutationActionDelete,
			Path:         m.Path,
			OriginalBody: m.Body,
		}
	case MutationActionUpdate:
		return &Mutation{
			ClientID:     invertID(m.ClientID),
			ID:           invertID(m.ID),
			Timestamp:    timeNow(),
			Conn:         m.Conn,
			Action:       MutationActionUpdate,
			Path:         m.Path,
			OriginalBody: m.Body,
			Body:         m.OriginalBody,
		}
	case MutationActionDelete:
		return &Mutation{
			ClientID:  invertID(m.ClientID),
			ID:        invertID(m.ID),
			Timestamp: timeNow(),
			Conn:      m.Conn,
			Action:    MutationActionCreate,
			Path:      m.Path,
			Body:      m.OriginalBody,
		}
	}

	return nil
}

// Equivalent returns true if the two mutations are equivalent, but not
// necessarily identical.
//
// Equivalent mutations are mutations that would have the same effect if
// applied. For example, two mutations that update the same path with the same
// value are equivalent. It does not evaluate the original body of the mutation
// or the mutation ID. This means that the inverse of two equivalent mutations
// are not necessarily equivalent.
func (m *Mutation) Equivalent(other *Mutation) bool {
	if other == nil {
		return false
	}

	if m.Action != other.Action {
		return false
	}

	if len(m.Path) != len(other.Path) {
		return false
	}

	for i, p := range m.Path {
		if p != other.Path[i] {
			return false
		}
	}

	return bytes.Equal(m.Body, other.Body)
}

// Equal returns true if the two mutations are identical.
func (m *Mutation) Equal(other *Mutation) bool {
	return m.Equivalent(other) &&
		m.ID == other.ID &&
		m.ClientID == other.ClientID &&
		bytes.Equal(m.OriginalBody, other.OriginalBody) &&
		m.Timestamp.Equal(other.Timestamp) &&
		m.Conn == other.Conn
}

// ParseMutation parses a mutation from a string.
//
// The string should be in the following format:
// [timestamp] <id> <action> <path> [body]
func ParseMutation(msg string, conn Conn) (*Mutation, error) {
	split := strings.Split(msg, " ")
	if len(split) < 3 { // must have at least id, action and path
		return nil, &ErrMutationFailed{Msg: "invalid mutation"}
	}

	timestamp, err := time.Parse(time.RFC3339, split[0])
	if err != nil {
		timestamp = timeNow()
	} else {
		split = split[1:]
	}

	var path []string
	if split[2] == "/" {
		path = []string{}
	} else {
		path = strings.Split(split[2], "/")
		if path[0] == "" {
			path = path[1:]
		}
	}

	var body []byte
	if len(split) > 3 {
		body = []byte(strings.Join(split[3:], " "))
	}

	return &Mutation{
		ClientID:  MutationID(split[0]),
		ID:        generateMutationID(),
		Timestamp: timestamp,
		Conn:      conn,
		Action:    ParseMutationAction(split[1]),
		Path:      path,
		Body:      body,
	}, nil
}
