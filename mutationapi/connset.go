package mutationapi

import (
	"context"
	"sync"
)

// ConnSet is a set of connections.
type ConnSet struct {
	values map[Conn]struct{}
	lock   sync.RWMutex
}

// Add adds a connection to the set.
func (s *ConnSet) Add(conn Conn) {
	s.lock.Lock()
	s.values[conn] = struct{}{}
	s.lock.Unlock()
}

// Remove removes a connection from the set.
func (s *ConnSet) Remove(conn Conn) {
	s.lock.Lock()
	delete(s.values, conn)
	s.lock.Unlock()
}

// Broadcast sends a mutation to all connections in the set.
func (s *ConnSet) Broadcast(mutation *Mutation) {
	// It might make sense to read all the connections first, then unlock, then send
	// the mutations.
	s.lock.RLock()
	for conn := range s.values {
		conn.Send(mutation)
	}
	s.lock.RUnlock()
}

// Purge removes all closed connections from the set.
func (s *ConnSet) Purge() {
	s.lock.Lock()
	for conn := range s.values {
		if conn.IsClosed() {
			s.Remove(conn)
		}
	}
	s.lock.Unlock()
}

// PipeAll receives mutations from all connections in a ConnSet and sends them to
// a channel until the ConnSet is closed or the provided context is done.
func (s *ConnSet) PipeAll(ctx context.Context, mutations chan<- *Mutation) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for conn := range s.values {
		go Pipe(ctx, conn, mutations)
	}
}

// NewConnSet creates a new empty connection set.
func NewConnSet() *ConnSet {
	return &ConnSet{
		values: make(map[Conn]struct{}),
	}
}
