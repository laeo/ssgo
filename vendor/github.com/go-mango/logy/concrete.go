package logy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type logy struct {
	*sync.Mutex
	name       string
	writer     *tabwriter.Writer
	funnel     int
	dateFormat string
}

// SetOutput sets messages should be writed into.
func (l *logy) SetOutput(w io.Writer) {
	l.writer = tabwriter.NewWriter(w, 8, 8, 1, '\t', tabwriter.AlignRight|tabwriter.TabIndent)
}

// SetWriteLevel sets minimal level that should be write actually.
func (l *logy) SetWriteLevel(lv int) {
	l.funnel = lv
}

// SetDateFormat sets date format of message creation time.
// RTFM https://golang.org/src/time/format.go
func (l *logy) SetDateFormat(f string) {
	l.dateFormat = f
}

// SetName set logy name.
func (l *logy) SetName(name string) {
	l.name = name
}

// Clone returns cloned logy struct.
func (l *logy) Clone() Logy {
	n := *l
	return &n
}

func (l *logy) print(lv int, s ...string) {
	if lv >= l.funnel {
		l.Lock()
		fmt.Fprintf(
			l.writer,
			"%s  %s  %s\t%s\n",
			colorful(colorCyan, l.name),
			time.Now().Format(l.dateFormat),
			colorful(colors[lv], levels[lv]),
			strings.Join(s, " "),
		)

		l.writer.Flush()
		l.Unlock()
	}
}

func (l *logy) printf(lv int, s string, a ...interface{}) {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 8, 8, 1, '\t', tabwriter.TabIndent)
	fmt.Fprintf(w, s, a...)
	w.Flush()

	l.print(lv, string(b.Bytes()))
}

// Debug put a message to logy.
func (l *logy) Debug(s ...string) {
	l.print(levelDebug, s...)
}

// Debugf put a message with given format to logy.
func (l *logy) Debugf(s string, a ...interface{}) {
	l.printf(levelDebug, s, a...)
}

// Info put a message to logy.
func (l *logy) Info(s ...string) {
	l.print(levelInfo, s...)
}

// Infof put a message to logy.
func (l *logy) Infof(s string, a ...interface{}) {
	l.printf(levelInfo, s, a...)
}

// Warn put a message to logy.
func (l *logy) Warn(s ...string) {
	l.print(levelWarn, s...)
}

// Warnf put a message to logy.
func (l *logy) Warnf(s string, a ...interface{}) {
	l.printf(levelWarn, s, a...)
}

// Error put a message to logy, then shutdown the process with exit code 1.
func (l *logy) Error(s ...string) {
	l.print(levelError, s...)
	os.Exit(1)
}

// Errorf put a message to logy, then shutdown the process with exit code 1.
func (l *logy) Errorf(s string, a ...interface{}) {
	l.printf(levelError, s, a...)
	os.Exit(1)
}
