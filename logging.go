package mockaso

import (
	"context"
	"fmt"
	"log"
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

// SlogLogger implementation of Logger using an slog.Logger.
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

// LogLogger implementation of Logger using a log.Logger.
type LogLogger struct {
	logger *log.Logger
}

func (l *LogLogger) Log(args ...any) {
	l.logger.Println(fmt.Sprint(args...))
}

func (l *LogLogger) Logf(format string, args ...any) {
	l.logger.Printf(format, args...)
}

func NewLogLogger(logger *log.Logger) *LogLogger {
	return &LogLogger{logger: logger}
}
