package mutationapi

// Conn something that can send and receive mutations. Connections can be closed
// by calling Close. Once a connection is closed, it cannot be used again. If a
// connection is closed, any calls to Send or Receive will return ErrClosed.
//
// Connections may close themselves based on external events, such as a
// websocket connection being closed by the client.
//
// Conn implementations must be safe for concurrent use.
type Conn interface {
	// Send sends a mutation to the connection. If the connection is closed, this
	// will return ErrClosed.
	Send(*Mutation) error

	// Receive receives a mutation from the connection. If the connection is
	// closed, this will return ErrClosed.
	Receive() (*Mutation, error)

	// Close closes the connection. Once a connection is closed, it cannot be used
	// again. If a connection is closed, any calls to Send or Receive will return
	// ErrClosed.
	Close() error

	// IsClosed returns true if the connection is closed.
	IsClosed() bool

	// sendError sends an error to the connection. Different connection types may
	// handle errors differently.
	sendError(MutationID, error)
}

// ConnHandlerFunc is a function that accepts new Conn objects from a Conn
// generator.
type ConnHandlerFunc func(Conn)
