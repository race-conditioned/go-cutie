// Package cutie is a minimal, pretty structured logger for Go.
//
// Inspired by Go's log/slog. Two handlers, one logger, one banner.
//
//	log := cutie.New(new(cutie.PrettyHandler))
//	log.Info("server started", cutie.Attrs{"port": 8080})
//
//	cutie.PrintBannerPick("my-app", cfg, []string{"port", "stage"})
package cutie

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Output writers — defaults to os.Stdout/os.Stderr, swappable in tests.
var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// Level represents a log severity.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Attrs is a map of structured key-value pairs attached to a log record.
type Attrs = map[string]any

// LogRecord is the data passed from a Logger to a Handler.
type LogRecord struct {
	Level  Level
	Msg    string
	Time   time.Time
	Attrs  Attrs
	Expand *bool // nil = handler default, non-nil = override
}

// Handler processes a LogRecord.
type Handler interface {
	Handle(record LogRecord)
}

// Logger dispatches structured log records to a Handler.
type Logger struct {
	handler   Handler
	baseAttrs Attrs
	expand    *bool
}

// New creates a Logger backed by the given handler.
func New(handler Handler) *Logger {
	return &Logger{handler: handler}
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, attrs ...Attrs) {
	l.emit(LevelDebug, msg, attrs)
}

// Info logs at info level.
func (l *Logger) Info(msg string, attrs ...Attrs) {
	l.emit(LevelInfo, msg, attrs)
}

// Warn logs at warn level.
func (l *Logger) Warn(msg string, attrs ...Attrs) {
	l.emit(LevelWarn, msg, attrs)
}

// Error logs at error level.
func (l *Logger) Error(msg string, attrs ...Attrs) {
	l.emit(LevelError, msg, attrs)
}

// With returns a new Logger with the given attrs merged into every subsequent record.
// Does not mutate the original logger.
func (l *Logger) With(attrs Attrs) *Logger {
	merged := make(Attrs, len(l.baseAttrs)+len(attrs))
	for k, v := range l.baseAttrs {
		merged[k] = v
	}
	for k, v := range attrs {
		merged[k] = v
	}
	return &Logger{handler: l.handler, baseAttrs: merged, expand: l.expand}
}

// Expanded returns a Logger that forces expanded (multi-line) output for handlers that support it.
func (l *Logger) Expanded() *Logger {
	t := true
	return &Logger{handler: l.handler, baseAttrs: l.baseAttrs, expand: &t}
}

// Compact returns a Logger that forces compact (single-line) output for handlers that support it.
func (l *Logger) Compact() *Logger {
	f := false
	return &Logger{handler: l.handler, baseAttrs: l.baseAttrs, expand: &f}
}

func (l *Logger) emit(level Level, msg string, attrs []Attrs) {
	merged := make(Attrs, len(l.baseAttrs))
	for k, v := range l.baseAttrs {
		merged[k] = v
	}
	if len(attrs) > 0 && attrs[0] != nil {
		for k, v := range attrs[0] {
			merged[k] = v
		}
	}
	l.handler.Handle(LogRecord{
		Level:  level,
		Msg:    msg,
		Time:   time.Now(),
		Attrs:  merged,
		Expand: l.expand,
	})
}

// writerForLevel returns stdout for debug/info, stderr for warn/error.
func writerForLevel(level Level) io.Writer {
	if level == LevelWarn || level == LevelError {
		return stderr
	}
	return stdout
}

// ptrBool returns a pointer to a bool value.
func ptrBool(b bool) *bool {
	return &b
}

// fprintln writes a line to the given writer.
func fprintln(w io.Writer, s string) {
	fmt.Fprintln(w, s)
}
