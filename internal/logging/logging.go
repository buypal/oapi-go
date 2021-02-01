package logging

import (
	"fmt"
	"strings"
)

// Level type, taken from logrus
type Level uint32

// These are the different logging levels. You can set the logging level to log
// on your instance of logger, obtained with `logrus.New()`.
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// ParseLevel takes a string level and returns the Logrus log level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid logrus Level: %q", lvl)
}

// Printer simple interface for logging
type Printer interface {
	Printf(Level, string, ...interface{})
}

type loggingFn func(Level, string, ...interface{})

func (fn loggingFn) Printf(lvl Level, m string, args ...interface{}) {
	if fn == nil {
		return
	}
	fn(lvl, m, args...)
}

// Void returns suppressed logger
func Void() Printer {
	return (loggingFn)(nil)
}

// NewLogger returns new logger
func NewLogger(fn loggingFn) Printer {
	return fn
}

// Log will log whatever, behaves sames as printf
func Log(log Printer, lvl Level, m string, args ...interface{}) {
	log.Printf(lvl, m, args...)
}

// Debug will log with debug level
func Debug(log Printer, m string, args ...interface{}) {
	log.Printf(DebugLevel, m, args...)
}

// Fatal will log with fatal level
func Fatal(log Printer, m string, args ...interface{}) {
	log.Printf(FatalLevel, m, args...)
}

// Warn will log with want level
func Warn(log Printer, m string, args ...interface{}) {
	log.Printf(WarnLevel, m, args...)
}

// Info will log with info level
func Info(log Printer, m string, args ...interface{}) {
	log.Printf(InfoLevel, m, args...)
}
