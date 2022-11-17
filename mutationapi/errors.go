package mutationapi

import (
	"fmt"
)

type ErrCommunicationFailed struct {
	Err  error
	Msg  string
	Conn Conn
}

func (e *ErrCommunicationFailed) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("communication failed: %v: %v", e.Msg, e.Err)
	}

	return fmt.Sprintf("communication failed: %v", e.Msg)
}

func (e *ErrCommunicationFailed) Unwrap() error {
	return e.Err
}

func (e *ErrCommunicationFailed) Is(target error) bool {
	_, ok := target.(*ErrCommunicationFailed)
	return ok
}

type ErrMutationFailed struct {
	Err error
	Msg string
	Mut *Mutation
}

func (e *ErrMutationFailed) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("mutation failed: %v: %v", e.Msg, e.Err)
	}
	return fmt.Sprintf("mutation failed: %v", e.Msg)
}

func (e *ErrMutationFailed) Unwrap() error {
	return e.Err
}

func (e *ErrMutationFailed) Is(target error) bool {
	_, ok := target.(*ErrMutationFailed)
	return ok
}

type ErrClosed struct {
	Err  error
	Conn Conn
}

func (e *ErrClosed) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("connection %v closed: %v", e.Conn, e.Err)
	}
	return fmt.Sprintf("connection %v closed", e.Conn)
}

func (e *ErrClosed) Unwrap() error {
	return e.Err
}

func (e *ErrClosed) Is(target error) bool {
	_, ok := target.(*ErrClosed)
	return ok
}

type ErrInvalidFilterRule struct {
	Rule string
}

func (e *ErrInvalidFilterRule) Error() string {
	return fmt.Sprintf("invalid filter rule: %v", e.Rule)
}

func (e *ErrInvalidFilterRule) Is(target error) bool {
	_, ok := target.(*ErrInvalidFilterRule)
	return ok
}
