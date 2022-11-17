package radiotracker

import (
	"bufio"
	"context"
	"time"

	"go.bug.st/serial"
)

type rawMessage struct {
	raw string
	ts  time.Time
}

type RadioTracker struct {
	PortName string
	Baud     int

	parentCTX context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	port serial.Port

	rawMessages chan *rawMessage
	messages    chan *Message
	errors      chan error
}

func (r *RadioTracker) getRawMessages() {
	scanner := bufio.NewScanner(r.port)

	for scanner.Scan() {
		r.rawMessages <- &rawMessage{
			raw: scanner.Text(),
			ts:  time.Now(),
		}
	}

	if err := scanner.Err(); err != nil {
		r.errors <- err
	}
}

func (r *RadioTracker) handleMessages() {
	for {
		select {
		case rawMsg := <-r.rawMessages:
			msg, err := ParseMessage(rawMsg)
			if err != nil {
				r.errors <- err
				continue
			}

			r.messages <- msg
		case <-r.ctx.Done():
			r.port.Close()
		}
	}
}

func (r *RadioTracker) Messages() <-chan *Message {
	return r.messages
}

func (r *RadioTracker) Listen() error {
	ctx, cancel := context.WithCancel(r.parentCTX)

	r.ctx = ctx
	r.cancel = cancel

	mode := &serial.Mode{
		BaudRate: r.Baud,
	}

	port, err := serial.Open(r.PortName, mode)
	if err != nil {
		return err
	}

	r.port = port

	go r.getRawMessages()
	go r.handleMessages()

	return nil
}

func (r *RadioTracker) Errors() <-chan error {
	return r.errors
}

func (r *RadioTracker) Close() {
	r.cancel()
}

func New(ctx context.Context) *RadioTracker {
	r := &RadioTracker{
		Baud: 9600,

		parentCTX: ctx,

		errors:      make(chan error),
		rawMessages: make(chan *rawMessage, 16),
		messages:    make(chan *Message, 16),
	}

	availablePorts, err := serial.GetPortsList()
	if err == nil {
		r.PortName = availablePorts[0]
	}

	return r
}
