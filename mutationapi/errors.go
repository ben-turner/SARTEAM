package mutationapi

import "errors"

var (
	// ErrClosed is returned when reading from or writing to closed connection.
	ErrClosed = errors.New("connection closed")
	// ErrInvalidArgument is returned when an invalid argument is passed to a function.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrInvalidFilterRule is returned when an invalid filter rule is used.
	ErrInvalidFilterRule = errors.New("invalid filter rule")
)
