package mutationapi

import (
	"errors"
	"testing"
)

var defaultErrorLogger = errorLogger
var defaultInfoLogger = infoLogger

func TestDefaultLoggers(t *testing.T) {
	t.Parallel()

	// The default loggers are noop functions. We call them here for coverage and
	// to make sure they aren't nil.
	defaultErrorLogger(errors.New("test error"))
	defaultInfoLogger("Hello World")
}

func TestLoggerSetting(t *testing.T) {
	infoBefore := infoLogger
	errorBefore := errorLogger

	defer func() {
		infoLogger = infoBefore
		errorLogger = errorBefore
	}()

	SetInfoLogger(nil)
	SetErrorLogger(nil)

	if infoLogger == nil {
		t.Fatal("Unexpected nil info logger")
	}

	if errorLogger == nil {
		t.Fatal("Unexpected nil error logger")
	}

	var infoMsg string
	testInfoLogger := func(msg string) {
		infoMsg = msg
	}

	var loggedErr error
	testErrorLogger := func(err error) {
		loggedErr = err
	}

	SetInfoLogger(testInfoLogger)
	SetErrorLogger(testErrorLogger)

	testErr := errors.New("test error")
	infoLogger("Hello World")
	errorLogger(testErr)

	if infoMsg != "Hello World" {
		t.Fatal("Unexpected info message")
	}

	if loggedErr != testErr {
		t.Fatal("Unexpected error")
	}

	infoLogger = infoBefore
	errorLogger = errorBefore
}
