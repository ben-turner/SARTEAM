package mutationapi

import (
	"errors"
	"testing"
)

func TestErrCommunicationFailed(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	wrapped := errors.New("wrapped error")
	ecf := &ErrCommunicationFailed{
		wrapped,
		"test message",
		conn,
	}
	err := error(ecf)

	if !errors.Is(err, wrapped) {
		t.Fatal("Unexpected error")
	}

	if !errors.Is(err, &ErrCommunicationFailed{}) {
		t.Fatal("Unexpected error")
	}

	if ecf.Unwrap() != wrapped {
		t.Fatal("Unexpected error unwrap")
	}

	if ecf.Error() != "communication failed: test message: wrapped error" {
		t.Fatal("Unexpected error message")
	}

	ecf.Err = nil

	if ecf.Error() != "communication failed: test message" {
		t.Fatal("Unexpected error message")
	}
}

func TestErrMutationFailed(t *testing.T) {
	t.Parallel()

	wrapped := errors.New("wrapped error")
	emf := &ErrMutationFailed{
		wrapped,
		"test message",
		&Mutation{},
	}
	err := error(emf)

	if !errors.Is(err, wrapped) {
		t.Fatal("Unexpected error")
	}

	if !errors.Is(err, &ErrMutationFailed{}) {
		t.Fatal("Unexpected error")
	}

	if emf.Unwrap() != wrapped {
		t.Fatal("Unexpected error unwrap")
	}

	if emf.Error() != "mutation failed: test message: wrapped error" {
		t.Fatal("Unexpected error message")
	}

	emf.Err = nil

	if emf.Error() != "mutation failed: test message" {
		t.Fatal("Unexpected error message")
	}
}

func TestErrClosed(t *testing.T) {
	t.Parallel()

	conn := NoopConn()
	defer conn.Close()

	wrapped := errors.New("wrapped error")
	ec := &ErrClosed{
		wrapped,
		conn,
	}
	err := error(ec)

	if !errors.Is(err, wrapped) {
		t.Fatal("Unexpected error")
	}

	if !errors.Is(err, &ErrClosed{}) {
		t.Fatal("Unexpected error")
	}

	if ec.Unwrap() != wrapped {
		t.Fatal("Unexpected error unwrap")
	}

	if ec.Error() != "connection NoopConn closed: wrapped error" {
		t.Fatal("Unexpected error message")
	}

	ec.Err = nil

	if ec.Error() != "connection NoopConn closed" {
		t.Fatal("Unexpected error message")
	}
}

func TestErrInvalidFilterRule(t *testing.T) {
	t.Parallel()

	rule := "!"

	eifr := &ErrInvalidFilterRule{rule}
	if eifr.Error() != "invalid filter rule: !" {
		t.Fatal("Unexpected error message")
	}

	if !errors.Is(error(eifr), &ErrInvalidFilterRule{}) {
		t.Fatal("Unexpected error")
	}
}
