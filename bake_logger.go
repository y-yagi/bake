package main

import (
	"fmt"
	"io"

	"github.com/y-yagi/color"
)

// BakeLogger is a logger for bake.
type BakeLogger struct {
	w io.Writer
}

var (
	green = color.New(color.FgGreen, color.Bold).SprintFunc()
)

// NewLogger creates a new BakeLogger.
func NewLogger(w io.Writer) *BakeLogger {
	l := &BakeLogger{w: w}
	return l
}

// Printf print log with format.
func (l *BakeLogger) Printf(action, format string, a ...interface{}) {
	var log string

	if len(action) != 0 {
		log += fmt.Sprintf("%s ", green(action))
	}
	log += fmt.Sprintf(format, a...)
	fmt.Fprint(l.w, log)
}
