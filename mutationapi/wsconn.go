package mutationapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// wsMessage is a message sent to or received from a websocket connection.
type wsMessage struct {
	t      int
	p      []byte
	result chan error
}

// wsConn is a websocket connection. It maintains the websocket connection, and
// handles reading and writing mutations.
type wsConn struct {
	ws *websocket.Conn

	ctx    context.Context
	cancel context.CancelFunc

	recMessages  chan *wsMessage
	sendMessages chan *wsMessage
	mutations    chan *Mutation

	pingTicker    *time.Ticker
	closedTimer   *time.Timer
	closedTimeout time.Duration
}

// read is a goroutine that receives messages from the websocket and sends them
// to the recMessages channel. It is expected that read is the only goroutine
// reading from the websocket. read must be called exactly once.
func (c *wsConn) read() {
	for {
		t, p, err := c.ws.ReadMessage()
		if err != nil {
			c.cancel()
			return
		}

		c.recMessages <- &wsMessage{
			t: t,
			p: p,
		}
	}
}

// writeUnsafe sends a message to the websocket directly. writeUnsafe is not
// thread-safe and should only be called from the write() goroutine.
func (c *wsConn) writeUnsafe(msg *wsMessage) error {
	if c.ws == nil {
		return ErrClosed
	}

	switch msg.t {
	case websocket.TextMessage, websocket.BinaryMessage:
		return c.ws.WriteMessage(msg.t, msg.p)
	case websocket.CloseMessage, websocket.PingMessage, websocket.PongMessage:
		return c.ws.WriteControl(msg.t, msg.p, time.Now().Add(10*time.Second))
	}

	return ErrInvalidArgument
}

// write is a goroutine that sends messages from the sendMessages channel to the
// websocket. It is expected that write is the only goroutine writing to the
// websocket. write must be called exactly once.
func (c *wsConn) write() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.sendMessages:
			err := c.writeUnsafe(msg)
			msg.result <- err
		}
	}
}

// work is a goroutine that processes messages and manages the connection.
func (c *wsConn) work() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.closedTimer.C:
			c.Close()
		case <-c.pingTicker.C:
			c.sendMessages <- &wsMessage{t: websocket.PingMessage}
		case msg := <-c.recMessages:
			c.closedTimer.Stop()
			c.closedTimer.Reset(c.closedTimeout)
			if msg.t == websocket.TextMessage {
				mut := ParseMutation(string(msg.p), c)
				c.mutations <- mut
			}
		}
	}
}

// Send sends a mutation to the websocket. If the connection is closed, Send
// will return ErrClosed.
func (c *wsConn) Send(mut *Mutation) error {
	res := make(chan error)
	c.sendMessages <- &wsMessage{
		t:      websocket.TextMessage,
		p:      []byte(mut.String()),
		result: res,
	}

	return <-res
}

// Receive receives a mutation from the websocket. If the connection is closed,
// Receive will return ErrClosed.
func (c *wsConn) Receive() (*Mutation, error) {
	select {
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	case mut := <-c.mutations:
		return mut, nil
	}
}

// Close closes the websocket connection.
func (c *wsConn) Close() error {
	c.cancel()
	err := c.ws.Close()
	c.ws = nil
	return err
}

// IsClosed returns true if the connection is closed.
func (c *wsConn) IsClosed() bool {
	return c.ws == nil
}

// sendError sends an error to the websocket with the mutation ID that caused
// the error. If the connection is closed, sendError will fail silently.
func (c *wsConn) sendError(originator MutationID, err error) {
	msg := fmt.Sprintf("%s ERROR %s", originator, err)
	c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

// WebsocketHandler is a handler that accepts http requests, upgrades them to
// websocket connections, and creates a Conn for each connection. Each Conn is
// passed to the handler function.
type WebsocketHandler struct {
	// Handler is the function that will be called for each new Conn.
	Handler ConnHandlerFunc
	// Upgrader is the websocket upgrader that will be used to upgrade the
	// http request to a websocket connection.
	Upgrader websocket.Upgrader
	// ResponseHeader is the http header that will be sent in the response to
	// the http request.
	ResponseHeader http.Header

	// PingInterval is the interval at which pings will be sent to the websocket
	// connection. If PingInterval is zero any calls to ServeHTTP will panic. It
	// is recommended that PingInterval be set to a value less than Timeout.
	PingInterval time.Duration
	// Timeout is the amount of time that the connection will wait for a message
	// before closing. If Timeout is 0, any calls to ServeHTTP will panic. It is
	// recommended that Timeout be set to a multiple of PingInterval.
	Timeout time.Duration
}

// ServeHTTP upgrades the http request to a websocket connection and creates a
// Conn for it. The Conn is passed to the handler function.
func (h *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := h.Upgrader.Upgrade(w, r, h.ResponseHeader)
	if err != nil {
		// TODO: Handle this... somehow
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	conn := &wsConn{
		ws:           ws,
		ctx:          ctx,
		cancel:       cancel,
		recMessages:  make(chan *wsMessage),
		sendMessages: make(chan *wsMessage),
		mutations:    make(chan *Mutation),

		pingTicker:    time.NewTicker(h.PingInterval),
		closedTimeout: h.Timeout,
		closedTimer:   time.NewTimer(h.Timeout),
	}

	go conn.read()
	go conn.write()
	go conn.work()

	h.Handler(conn)
}
