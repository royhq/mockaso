package mockaso

import (
	"context"
	"fmt"
	"log/slog"
)

// Logger abstraction intended for use with testing.T.
type Logger interface {
	Log(...any)
	Logf(string, ...any)
}

// noLogger is a Logger that does not log anything.
type noLogger struct{}

func (n noLogger) Log(...any)          {}
func (n noLogger) Logf(string, ...any) {}

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

func NewSlogLogger(logger *slog.Logger, level slog.Level) *SlogLogger {
	return &SlogLogger{logger: logger, level: level}
}
