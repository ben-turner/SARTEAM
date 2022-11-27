package mutationapi

import (
	"context"
	"errors"
)

type ConnMock struct {
	Ctx                context.Context
	Cancel             context.CancelFunc
	SendChan           chan *Mutation
	ErrorChan          chan error
	ReceivableMutation *Mutation
	Name               string
}

// Send sends a mutation to the connection. If the connection is closed, this
// will return ErrClosed.
func (c *ConnMock) Send(m *Mutation) error {
	if c.SendChan == nil {
		return &ErrClosed{
			Err:  errors.New("Stubbed connection cannot send"),
			Conn: c,
		}
	}

	c.SendChan <- m

	return nil
}

// Receive receives a mutation from the connection. If the connection is
// closed, this will return ErrClosed.
func (c *ConnMock) Receive() (*Mutation, error) {
	if c.ReceivableMutation == nil {
		return nil, &ErrClosed{
			Err:  errors.New("Stubbed connection cannot receive"),
			Conn: c,
		}
	}

	return c.ReceivableMutation, nil
}

// Close closes the connection. Once a connection is closed, it cannot be used
// again. If a connection is closed, any calls to Send or Receive will return
// ErrClosed.
func (c *ConnMock) Close() error {
	c.SendChan = nil
	c.ReceivableMutation = nil
	c.Cancel()

	return nil
}

// IsClosed returns true if the connection is closed.
func (c *ConnMock) IsClosed() bool {
	return c.SendChan == nil && c.ReceivableMutation == nil
}

// String returns a string representation of the mutation. This should be a
// human-readable representation of the mutation.
func (c *ConnMock) String() string {
	return c.Name
}

// sendError sends an error to the connection. Different connection types may
// handle errors differently.
func (c *ConnMock) sendError(err error) {
	if c.ErrorChan == nil {
		return
	}

	c.ErrorChan <- err
}

func NoopConn() *ConnMock {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConnMock{
		Ctx:                ctx,
		Cancel:             cancel,
		SendChan:           BlackholeChanWithContext[*Mutation](ctx),
		ErrorChan:          BlackholeChanWithContext[error](ctx),
		ReceivableMutation: nil,
		Name:               "NoopConn",
	}
}

func ConnStub(sendChan chan *Mutation, errorChan chan error) *ConnMock {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConnMock{
		Ctx:                ctx,
		Cancel:             cancel,
		SendChan:           sendChan,
		ErrorChan:          errorChan,
		ReceivableMutation: nil,
		Name:               "ConnStub",
	}
}

func BlackholeChanWithContext[T any](ctx context.Context) chan T {
	c := make(chan T)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c:
				// Do nothing
			}
		}
	}()

	return c
}

func BlackholeChan[T any]() (chan T, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	return BlackholeChanWithContext[T](ctx), cancel
}
