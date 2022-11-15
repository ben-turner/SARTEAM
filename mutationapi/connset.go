package mutationapi

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
