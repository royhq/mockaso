package mockaso_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/royhq/mockaso"
)

func TestServer(t *testing.T) {
	t.Parallel()

	server := mockaso.NewServer(mockaso.WithLogger(t))

	t.Run("before start", func(t *testing.T) {
		t.Run("should not return inner test server", func(t *testing.T) {
			assert.Nil(t, server.TestServer())
		})

		t.Run("should return empty URL", func(t *testing.T) {
			assert.Zero(t, server.URL())
		})

		t.Run("should do nothing when shutdown", func(t *testing.T) {
			assert.NoError(t, server.Shutdown())
		})

		t.Run("should not panic when clear", func(t *testing.T) {
			assert.NotPanics(t, server.Clear)
		})
	})

	t.Run("start (server started)", func(t *testing.T) {
		t.Run("should start server", func(t *testing.T) {
			assert.NoError(t, server.Start())
		})

		t.Run("should not return error when started", func(t *testing.T) {
			assert.NoError(t, server.Start())
		})

		t.Run("should return inner test server", func(t *testing.T) {
			assert.NotNil(t, server.TestServer())
		})

		t.Run("should not panic when clear", func(t *testing.T) {
			assert.NotPanics(t, server.Clear)
		})

		t.Run("should return URL", func(t *testing.T) {
			innerServer := server.TestServer()
			assert.Equal(t, innerServer.URL, server.URL())
		})
	})

	t.Run("shutdown (server closed)", func(t *testing.T) {
		t.Run("should shutdown server", func(t *testing.T) {
			assert.NoError(t, server.Shutdown())
		})

		t.Run("should shutdown server when is already closed", func(t *testing.T) {
			assert.NoError(t, server.Shutdown())
		})

		t.Run("should not panic when clear", func(t *testing.T) {
			assert.NotPanics(t, server.Clear)
		})
	})
}
