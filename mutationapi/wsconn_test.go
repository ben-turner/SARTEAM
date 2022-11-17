package mutationapi

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type stubMessage struct {
	messageType int
	message     []byte
}

type StubWS struct {
	closed       bool
	sentMessages chan stubMessage
	receivable   []stubMessage
	nextError    error
}

func (c *StubWS) ReadMessage() (int, []byte, error) {
	if c.nextError != nil {
		err := c.nextError
		c.nextError = nil
		return 0, nil, err
	}

	if len(c.receivable) == 0 {
		return 0, nil, errors.New("no more messages to receive")
	}

	msg := c.receivable[0]
	c.receivable = c.receivable[1:]

	return msg.messageType, msg.message, nil
}

func (c *StubWS) WriteMessage(messageType int, data []byte) error {
	if c.nextError != nil {
		err := c.nextError
		c.nextError = nil
		return err
	}

	c.sentMessages <- stubMessage{
		messageType: messageType,
		message:     data,
	}

	return nil
}

func (c *StubWS) WriteControl(messageType int, data []byte, _ time.Time) error {
	if c.nextError != nil {
		err := c.nextError
		c.nextError = nil
		return err
	}

	c.sentMessages <- stubMessage{
		messageType: messageType,
		message:     data,
	}

	return nil
}

func (c *StubWS) Close() error {
	if c.nextError != nil {
		err := c.nextError
		c.nextError = nil
		return err
	}

	c.closed = true

	return nil
}

func (c *StubWS) RemoteAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func NewStubWS() *StubWS {
	return &StubWS{
		sentMessages: make(chan stubMessage, 1),
		receivable:   []stubMessage{},
	}
}

func newBaseWSConn() *wsConn {
	ctx, cancel := context.WithCancel(context.Background())

	return &wsConn{
		ws: NewStubWS(),

		ctx:    ctx,
		cancel: cancel,

		recMessages:  make(chan *wsMessage, 1),
		sendMessages: make(chan *wsMessage, 1),
		mutations:    make(chan *Mutation, 1),

		pingTicker:    time.NewTicker(1 * time.Second),
		closedTimer:   time.NewTimer(1 * time.Second),
		closedTimeout: 1 * time.Second,
	}
}

func TestWSClose(t *testing.T) {
	t.Parallel()

	conn := newBaseWSConn()
	ws := conn.ws.(*StubWS)

	if conn.IsClosed() {
		t.Fatal("connection should not be closed")
	}

	err := conn.Close()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if !conn.IsClosed() {
		t.Fatal("connection should be closed")
	}

	if conn.ws != nil || !ws.closed {
		t.Fatal("connection should be closed")
	}

	testErr := errors.New("test error")
	conn = newBaseWSConn()
	conn.ws.(*StubWS).nextError = testErr

	err = conn.Close()

	if err != testErr {
		t.Fatal("unexpected error:", err)
	}
}

func TestWSString(t *testing.T) {
	t.Parallel()

	c := newBaseWSConn()

	if c.String() != "127.0.0.1" {
		t.Error("unexpected address")
	}
}

func TestWSConnSendError(t *testing.T) {
	t.Parallel()

	conn := newBaseWSConn()

	testErr := errors.New("test error")

	conn.sendError(testErr)

	select {
	case msg := <-conn.ws.(*StubWS).sentMessages:
		if string(msg.message) != "ERROR test error" {
			t.Fatal("unexpected message:", msg)
		}

		if msg.messageType != websocket.TextMessage {
			t.Fatal("unexpected message type:", msg)
		}

	default:
		t.Fatal("message not sent")
	}

	testErr = &ErrMutationFailed{
		Err: errors.New("test error"),
		Msg: "test mutation",
		Mut: &Mutation{
			ID: "foobar",
		},
	}

	conn.sendError(testErr)

	select {
	case msg := <-conn.ws.(*StubWS).sentMessages:
		if string(msg.message) != "foobar ERROR mutation failed: test mutation: test error" {
			t.Fatal("unexpected message:", string(msg.message))
		}

		if msg.messageType != websocket.TextMessage {
			t.Fatal("unexpected message type:", msg)
		}

	default:
		t.Fatal("message not sent")
	}
}
