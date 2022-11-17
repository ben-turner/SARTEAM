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
	defer s.lock.Unlock()

	s.values[conn] = struct{}{}
}

// Remove removes a connection from the set.
func (s *ConnSet) Remove(conn Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.values, conn)
}

func (s *ConnSet) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.values)
}

// Broadcast sends a mutation to all connections in the set.
func (s *ConnSet) Broadcast(mutation *Mutation) {
	// It might make sense to read all the connections first, then unlock, then send
	// the mutations.
	s.lock.RLock()
	defer s.lock.RUnlock()

	for conn := range s.values {
		conn.Send(mutation)
	}
}

// Purge removes all closed connections from the set.
func (s *ConnSet) Purge() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for conn := range s.values {
		if conn.IsClosed() {
			delete(s.values, conn)
		}
	}
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
