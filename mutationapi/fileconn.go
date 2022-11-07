package mutationapi

import (
	"bufio"
	"os"
	"strconv"

	"github.com/ben-turner/sarteam/internal/singletons"
	"github.com/google/uuid"
)

type fileConn struct {
	file    *os.File
	scanner *bufio.Scanner
	line    int
	usedIDs map[MutationID]struct{}
}

// Send appends the mutation to the file.
func (c *fileConn) Send(mut *Mutation) error {
	// Replace the ID with a random UUID to ensure we don't get duplicates. UUIDs
	// might be a bit overkill but they're guaranteed to be unique. If file sizes
	// become unmangeable, using a shorter ID format might be worth considering.
	mut.ID = MutationID(uuid.New().String())
	c.usedIDs[mut.ID] = struct{}{}
	_, err := c.file.WriteString(mut.String())
	return err
}

// Receive reads the next mutation from the file.
func (c *fileConn) Receive() (*Mutation, error) {
start:
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}

	c.line++
	m := ParseMutation(c.scanner.Text(), c)
	if _, ok := c.usedIDs[m.ID]; ok {
		// Discard this as it's a duplicate, probably because we wrote it.
		// Use a goto to avoid excessive recursion and stack overflows for many
		// consecutive duplicates.
		goto start
	}

	c.usedIDs[m.ID] = struct{}{}

	// Replace the ID with the line number to make output more readable.
	m.ID = MutationID(strconv.Itoa(c.line))

	return m, nil
}

// Close closes the file.
func (c *fileConn) Close() error {
	err := c.file.Close()
	c.file = nil
	return err
}

// IsClosed returns true if the file is closed.
func (c *fileConn) IsClosed() bool {
	return c.file == nil
}

// sendError logs an error fith the filename and line number but does not store
// the error in the file.
func (c *fileConn) sendError(originator MutationID, err error) {
	// This logger should probably come from somewhere else.
	singletons.GetLogger().Printf("error in file %s:%d %s", c.file.Name(), originator, err)
}

// NewFileConn creates a new file connection.
//
// The file is read line by line and each line is parsed as a mutation. Writes
// are appended to the end of the file. Errors are logged but not stored in the
// file. In order to avoid reading mutations that were written by this instance,
// each mutation is assigned a random ID. This ID is replaced with the line
// number when the mutation is read.
//
// Returns *os.PathError if the file cannot be opened.
func NewFileConn(file string) (Conn, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &fileConn{
		file:    f,
		scanner: bufio.NewScanner(f),
		line:    0,
	}, nil
}
