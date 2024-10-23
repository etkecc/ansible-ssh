package logger

import (
	"log"
	"os"
)

var (
	withDebug bool
	logger    = log.New(os.Stdout, "[ansible-ssh] ", 0)
)

// Configure the logger package
func Configure(debug bool) {
	withDebug = debug
}

// Println logs the arguments to the standard logger.
func Println(args ...any) {
	logger.Println(args...)
}

// Debug logs the arguments to the standard logger if the debug flag is set.
func Debug(args ...any) {
	if !withDebug {
		return
	}
	logger.Println(args...)
}

// Fatal logs the arguments to the standard logger and then calls os.Exit(1).
func Fatal(args ...any) {
	logger.Fatal(args...)
}
