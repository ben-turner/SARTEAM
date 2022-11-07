package mutationapi

import "context"

// Pipe receives mutations from a *Conn and sends them to a channel until the
// *Conn is closed or the provided context is done. This is a blocking function.
func Pipe(ctx context.Context, conn Conn, mutations chan<- *Mutation) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			mutation, err := conn.Receive()
			if err != nil {
				return err
			}

			mutations <- mutation
		}
	}
}

// ConnSet is a set of connections.
type ConnSet map[Conn]struct{}

// Add adds a connection to the set.
func (s ConnSet) Add(conn Conn) {
	s[conn] = struct{}{}
}

// Remove removes a connection from the set.
func (s ConnSet) Remove(conn Conn) {
	delete(s, conn)
}

// Broadcast sends a mutation to all connections in the set.
func (s ConnSet) Broadcast(mutation *Mutation) {
	for conn := range s {
		conn.Send(mutation)
	}
}

// Purge removes all closed connections from the set.
func (s ConnSet) Purge() {
	for conn := range s {
		if conn.IsClosed() {
			s.Remove(conn)
		}
	}
}

// NewConnSet creates a new empty connection set.
func NewConnSet() ConnSet {
	return make(map[Conn]struct{})
}
