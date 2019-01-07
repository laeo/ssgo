package logy

import "io"

// Logy is a simple and useful logger package.
type Logy interface {
	SetOutput(io.Writer)
	SetWriteLevel(int)
	SetDateFormat(string)
	SetName(string)
	Clone() Logy

	// Debug prints all given string variables.
	Debug(...string)

	// Debugf prints all given string variales after formated by first argument.
	Debugf(string, ...interface{})

	// Info prints concated given string informations.
	Info(...string)

	// Infof prints concated given string informations after formated by first argument.
	Infof(string, ...interface{})
	Warn(...string)
	Warnf(string, ...interface{})

	// Error put a message to logy, then shutdown the process with exit code 1.
	Error(...string)

	// Errorf put a message to logy, then shutdown the process with exit code 1.
	Errorf(string, ...interface{})
}
