package mutationapi

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
)

// ioConn is a connection that reads and writes mutations to a file.
type ioConn struct {
	rwc     io.ReadWriteCloser
	scanner *bufio.Scanner
	line    int
	usedIDs map[MutationID]struct{}
	name    string
}

// Send appends the mutation to the file. The mutation is assigned a random ID
// to avoid reading mutations that were written by this instance.
func (c *ioConn) Send(mut *Mutation) error {
	mut.ID = generateMutationID() // This should probably not mutate the original mutation.
	c.usedIDs[mut.ID] = struct{}{}

	if _, err := io.WriteString(c.rwc, mut.String()+"\n"); err != nil {
		return &ErrCommunicationFailed{err, "failed to write", c}
	}

	return nil
}

// Receive reads the next mutation from the file. Invalid mutations are logged
// and skipped.
func (c *ioConn) Receive() (*Mutation, error) {
start:
	if !c.scanner.Scan() {
		err := c.scanner.Err()
		if errors.Is(err, os.ErrClosed) || err == nil {
			return nil, &ErrClosed{err, c}
		}
		return nil, &ErrCommunicationFailed{c.scanner.Err(), "failed to scan file", c}
	}

	c.line++
	m, err := ParseMutation(c.scanner.Text(), c)
	if err != nil {
		c.sendError(err)
		goto start
	}
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
func (c *ioConn) Close() error {
	err := c.rwc.Close()
	c.rwc = nil
	return err
}

// IsClosed returns true if the file is closed.
func (c *ioConn) IsClosed() bool {
	return c.rwc == nil
}

// String returns the file name.
func (c *ioConn) String() string {
	return c.name
}

// sendError logs an error fith the filename and line number but does not store
// the error in the file.
func (c *ioConn) sendError(err error) {
	errorLogger(err)
}

// NewIOConn creates a new Conn from a ReadWriteCloser.
//
// This is most useful for reading and writing mutations to files, but can be
// used for any ReadWriteCloser. If a file is used, it should be opened in
// append mode.
//
// Lines are read one by one from the ReadWriteCloser with each subsequent call
// to Receive. Invalid mutations are logged and skipped. Writes are appended to
// the end of the ReadWriteCloser. Errors are logged but written to the
// ReadWriteCloser.
//
// In order to avoid reading mutations that were written by this instance, each
// mutation is assigned a random ID when writing. This ID is replaced with the
// line number when the mutation is read.
//
// An optional name can be provided to identify the connection. In the case of
// files, this is the file name. If an empty string is provided, the string
// "unnamed IO connection" is used.
func NewIOConn(file io.ReadWriteCloser, name string) Conn {
	if name == "" {
		name = "unnamed IO connection"
	}

	return &ioConn{
		rwc:     file,
		name:    name,
		scanner: bufio.NewScanner(file),
		line:    0,
		usedIDs: make(map[MutationID]struct{}),
	}
}
