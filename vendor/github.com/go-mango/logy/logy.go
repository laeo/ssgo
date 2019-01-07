package logy

import (
	"os"
	"sync"
)

var std = New("LOGY")

// Std returns default Logy instance can be uses directly.
func Std() Logy {
	return std
}

// Clone returns cloned Logy struct.
func Clone(name string) Logy {
	logger := std.Clone()
	logger.SetName(name)
	return logger
}

// levels
const (
	levelDebug = iota
	levelInfo
	levelWarn
	levelError
)

// Colors
const (
	colorBlack = 30 + iota
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorPurple
	colorCyan
	colorWhite
	colorNone = 0
)

var (
	levels = []string{
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
	}

	colors = map[int]int{
		levelDebug: colorNone,
		levelInfo:  colorGreen,
		levelWarn:  colorYellow,
		levelError: colorRed,
	}
)

// New create and return new Logy instance.
func New(name string) Logy {
	l := &logy{
		new(sync.Mutex),
		name,
		nil,
		levelDebug,
		"15:04:05",
	}

	l.SetOutput(os.Stdout)

	return l
}
