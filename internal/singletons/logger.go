package singletons

import "log"

type Logger interface {
	Print(v ...any)
	Printf(format string, v ...any)
}

func GetLogger() Logger {
	return log.Default()
}
