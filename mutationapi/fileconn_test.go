package mutationapi

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

type testRWC struct {
	closed     bool
	data       []byte
	readCursor int
	nextErr    error
}

func (t *testRWC) Read(p []byte) (n int, err error) {
	if t.closed {
		return 0, os.ErrClosed
	}

	if t.nextErr != nil {
		err = t.nextErr
		t.nextErr = nil
		return
	}

	if t.readCursor >= len(t.data) {
		return 0, errors.New("EOF")
	}

	n = copy(p, t.data[t.readCursor:])
	t.readCursor += n

	return n, nil
}

func (t *testRWC) Write(p []byte) (n int, err error) {
	if t.closed {
		return 0, os.ErrClosed
	}

	if t.nextErr != nil {
		err = t.nextErr
		t.nextErr = nil
		return
	}

	t.data = append(t.data, p...)
	return len(p), nil
}

func (t *testRWC) Close() error {
	t.closed = true
	return nil
}

func NewTestRWC() *testRWC {
	return &testRWC{}
}

func TestIOConnSend(t *testing.T) {
	t.Parallel()

	rwc := &testRWC{}
	conn := NewIOConn(rwc, "test name")

	mut := &Mutation{
		ID:        "1",
		Timestamp: time.Date(2022, 11, 7, 14, 5, 10, 0, time.UTC),
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar"},
		body:      []byte("{}"),
	}

	conn.Send(mut)

	expected := "2022-11-07T14:05:10Z test-mutation-id CREATE foo/bar {}\n"
	if string(rwc.data) != expected {
		t.Fatalf("Expected %q, got %q", expected, rwc.data)
	}

	conn.Send(mut)

	testErr := errors.New("test error")
	rwc.nextErr = testErr

	err := conn.Send(mut)

	if !errors.Is(err, testErr) {
		t.Fatalf("Expected error %v, got %v", testErr, err)
	}

	if !errors.Is(err, &ErrCommunicationFailed{}) {
		t.Fatalf("Expected ErrCommunicationFailed, got %v", err)
	}
}

func TestIOConnReceive(t *testing.T) {
	// Not parallel because it uses the errorLogger global variable.
	errChan := make(chan error, 1)
	errorLogger = func(err error) {
		errChan <- err
	}

	testErr := errors.New("test error")

	suite := []struct {
		name       string
		msg        string
		errs       []error
		mut        *Mutation
		writeOnly  bool
		readErr    error
		loggedErrs []error
	}{
		{
			name: "no error",
			msg:  "2022-01-01T00:00:00Z foobar CREATE foo/bar {}",
			mut: &Mutation{
				ID:        "1",
				Timestamp: time.Date(2022, 11, 7, 14, 5, 10, 0, time.UTC),
				Action:    MutationActionCreate,
				Path:      []string{"foo", "bar"},
				body:      []byte("{}"),
			},
		},
		{
			name: "invalid mutation",
			msg:  "invalid mutation\n2022-01-01T00:00:00Z nextValid UPDATE foo/bar {\"foo\": \"bar\"}",
			loggedErrs: []error{
				&ErrMutationFailed{},
			},
			mut: &Mutation{
				ID:        "1",
				Timestamp: time.Date(2022, 11, 7, 14, 5, 10, 0, time.UTC),
				Action:    MutationActionUpdate,
				Path:      []string{"foo", "bar"},
				body:      []byte("{\"foo\": \"bar\"}"),
			},
		},
		{
			name:    "random error",
			readErr: testErr,
			errs:    []error{testErr, &ErrCommunicationFailed{}},
		},
		{
			name:    "ErrClosed",
			readErr: os.ErrClosed,
			errs:    []error{&ErrClosed{}},
		},
		{
			name:    "EOF",
			readErr: io.EOF,
			errs:    []error{&ErrClosed{}},
		},
	}

	for _, test := range suite {
		errChan = make(chan error, 1)

		fail := func(f string, args ...any) {
			t.Helper()
			msg := fmt.Sprintf(f, args...)
			t.Fatalf("Test %q failed: %v", test.name, msg)
		}

		rwc := &testRWC{}
		conn := NewIOConn(rwc, "test name")

		if test.msg != "" {
			fmt.Fprintln(rwc, test.msg)
		}

		if test.readErr != nil {
			rwc.nextErr = test.readErr
		}

		if test.writeOnly {
			continue
		}

		mut, err := conn.Receive()

		for _, expectedErr := range test.errs {
			if !errors.Is(err, expectedErr) {
				fail("expected error %v, got %v", expectedErr, err)
			}
		}

		if len(test.loggedErrs) > 0 {
			var err error
			select {
			case err = <-errChan:
			default:
				fail("expected error %v, got none", test.loggedErrs[0])
			}

			for _, expectedErr := range test.loggedErrs {
				if !errors.Is(err, expectedErr) {
					fail("expected error %v, got %v", expectedErr, err)
				}
			}
		}

		if mut != nil && !mut.Equivalent(test.mut) {
			fail("expected mutation %v, got %v", test.mut, mut)
		}
	}
}

func TestIOConnDuplicates(t *testing.T) {
	rwc := NewTestRWC()
	conn := NewIOConn(rwc, "test name")

	rwc.Write([]byte("2022-01-01T00:00:00Z foobar CREATE foo/bar {}\n"))
	rwc.Write([]byte("2022-01-01T00:00:00Z foobar CREATE foo/bar {}\n"))
	rwc.Write([]byte("2022-01-01T00:00:00Z notadupe UPDATE foo/bar {}\n"))

	mut, err := conn.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if mut == nil {
		t.Fatal("expected mutation, got nil")
	}
	if mut.ID != "1" {
		t.Fatalf("expected mutation ID 1, got %v", mut.ID)
	}
	if mut.Action != MutationActionCreate {
		t.Fatalf("expected mutation action CREATE, got %v", mut.Action)
	}

	mut, err = conn.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if mut == nil {
		t.Fatal("expected mutation, got nil")
	}
	if mut.ID != "3" {
		t.Fatalf("expected mutation ID 3, got %v", mut.ID)
	}
	if mut.Action != MutationActionUpdate {
		t.Fatalf("expected mutation action UPDATE, got %v", mut.Action)
	}

	conn.Send(&Mutation{
		ID:        "shouldbeignored",
		Timestamp: time.Date(2022, 11, 7, 14, 5, 10, 0, time.UTC),
		Action:    MutationActionCreate,
		Path:      []string{"foo", "bar"},
		body:      []byte("{}"),
	})
	rwc.Write([]byte("2022-01-01T00:00:00Z another DELETE foo/bar {}\n"))

	mut, err = conn.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if mut == nil {
		t.Fatal("expected mutation, got nil")
	}
	if mut.ID != "5" {
		t.Fatalf("expected mutation ID 5, got %v", mut.ID)
	}
	if mut.Action != MutationActionDelete {
		t.Fatalf("expected mutation action DELETE, got %v", mut.Action)
	}
}

func TestIOConnClose(t *testing.T) {
	t.Parallel()

	rwc := &testRWC{}

	conn := NewIOConn(rwc, "test name")
	if conn.IsClosed() {
		t.Fatal("Unexpected closed state")
	}

	conn.Close()

	if !conn.IsClosed() {
		t.Fatal("Expected closed state")
	}
}

func TestIOConnSendError(t *testing.T) {
	// Not parallel because it uses the errorLogger global variable.

	sentErr := errors.New("test error")
	errChan := make(chan error, 1)

	errorLogger = func(err error) {
		errChan <- err
	}

	rwc := &testRWC{}
	conn := NewIOConn(rwc, "test name")

	conn.sendError(sentErr)

	select {
	case receivedErr := <-errChan:
		if !errors.Is(receivedErr, sentErr) {
			t.Fatal("Unexpected error")
		}
	default:
		t.Fatal("Error not logged")
	}
}

func TestNewIOConnString(t *testing.T) {
	t.Parallel()

	rwc := &testRWC{}

	conn := NewIOConn(rwc, "test name")
	if conn.String() != "test name" {
		t.Fatal("Unexpected string")
	}

	conn = NewIOConn(rwc, "")
	if conn.String() != "unnamed IO connection" {
		t.Fatal("Unexpected string")
	}
}
