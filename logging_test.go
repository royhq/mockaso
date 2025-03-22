package mockaso_test

import (
	"bytes"
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/royhq/mockaso"
)

func TestSlogLogger(t *testing.T) {
	t.Parallel()

	t.Run("should Log", func(t *testing.T) {
		logger, buff := newTestSlogLogger(slog.LevelInfo)

		logger.Log("test message from an slog logger!!")

		logRegex := `time=[^ ]+ level=(\w+) msg="([^"]+)"`
		regex := regexp.MustCompile(logRegex)
		matches := regex.FindStringSubmatch(buff.String())

		assert.Len(t, matches, 3)
		assert.Equal(t, "INFO", matches[1])
		assert.Equal(t, "test message from an slog logger!!", matches[2])
	})

	t.Run("should Logf", func(t *testing.T) {
		logger, buff := newTestSlogLogger(slog.LevelWarn)

		logger.Logf("formated test message from %s!!", "an slog logger")

		logRegex := `time=[^ ]+ level=(\w+) msg="([^"]+)"`
		regex := regexp.MustCompile(logRegex)
		matches := regex.FindStringSubmatch(buff.String())

		assert.Len(t, matches, 3)
		assert.Equal(t, "WARN", matches[1])
		assert.Equal(t, "formated test message from an slog logger!!", matches[2])
	})
}

func newTestSlogLogger(level slog.Level) (*mockaso.SlogLogger, *bytes.Buffer) {
	var buff bytes.Buffer
	slogLogger := slog.New(slog.NewTextHandler(&buff, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := mockaso.NewSlogLogger(slogLogger, level)

	return logger, &buff
}
