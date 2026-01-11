package mockaso

import (
	"context"
	"fmt"
	"log"
	"log/slog"
)

// Logger abstraction intended for use with testing.T.
// The interface is compatible with testing.T's Log and Logf methods,
// allowing test loggers to be used directly with mock servers.
//
// Example with testing.T:
//
//	func TestAPI(t *testing.T) {
//		server := mockaso.NewServer(mockaso.WithLogger(t))
//		server.MustStart()
//		defer server.MustShutdown()
//		// Server will log to test output
//	}
type Logger interface {
	Log(...any)
	Logf(string, ...any)
}

// noLogger is a Logger that does not log anything.
type noLogger struct{}

func (n noLogger) Log(...any)          {}
func (n noLogger) Logf(string, ...any) {}

// SlogLogger implementation of Logger using an slog.Logger.
// All log messages are emitted at the configured level.
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	server := mockaso.NewServer(
//		mockaso.WithSlogLogger(logger, slog.LevelInfo),
//	)
type SlogLogger struct {
	logger *slog.Logger
	level  slog.Level
}

func (l *SlogLogger) Log(args ...any) {
	l.logger.Log(context.Background(), l.level, fmt.Sprint(args...))
}

func (l *SlogLogger) Logf(format string, args ...any) {
	l.logger.Log(context.Background(), l.level, fmt.Sprintf(format, args...))
}

// NewSlogLogger creates a new SlogLogger with the specified logger and level.
// All messages logged through this logger will use the given level.
func NewSlogLogger(logger *slog.Logger, level slog.Level) *SlogLogger {
	return &SlogLogger{logger: logger, level: level}
}

// LogLogger implementation of Logger using a log.Logger.
// This is useful for integration with standard library logging or legacy code.
//
// Example:
//
//	logger := log.New(os.Stdout, "[MOCKASO] ", log.LstdFlags)
//	server := mockaso.NewServer(mockaso.WithLogLogger(logger))
type LogLogger struct {
	logger *log.Logger
}

func (l *LogLogger) Log(args ...any) {
	l.logger.Println(fmt.Sprint(args...))
}

func (l *LogLogger) Logf(format string, args ...any) {
	l.logger.Printf(format, args...)
}

// NewLogLogger creates a new LogLogger wrapping the provided log.Logger.
func NewLogLogger(logger *log.Logger) *LogLogger {
	return &LogLogger{logger: logger}
}
