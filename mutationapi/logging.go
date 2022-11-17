package mutationapi

type ErrorLogger func(error)
type InfoLogger func(string)

var errorLogger ErrorLogger = func(err error) {
	// Do nothing.
}

var infoLogger InfoLogger = func(msg string) {
	// Do nothing.
}

func SetInfoLogger(logger InfoLogger) {
	if logger == nil {
		return
	}

	infoLogger = logger
}

func SetErrorLogger(logger ErrorLogger) {
	if logger == nil {
		return
	}

	errorLogger = logger
}
