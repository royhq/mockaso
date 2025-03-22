package mockaso_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/royhq/mockaso"
)

func TestServer(t *testing.T) {
	t.Parallel()

	server := mockaso.NewServer(mockaso.WithLogger(t))

	t.Run("before start", func(t *testing.T) {
		t.Run("should not return inner test server", func(t *testing.T) {
			assert.Nil(t, server.TestServer())
		})

		t.Run("should not return client", func(t *testing.T) {
			assert.Nil(t, server.Client())
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

		t.Run("should return client", func(t *testing.T) {
			assert.NotNil(t, server.Client())
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

func TestServer_Stub(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer()
	t.Cleanup(server.MustShutdown)

	client := server.Client()

	server.Stub(http.MethodGet, mockaso.URL("/api/users"))

	t.Run("should write response when match method and url", func(t *testing.T) {
		httpReq, _ := http.NewRequest(http.MethodGet, "/api/users", http.NoBody)

		httpResp, err := client.Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assert.Empty(t, httpResp.Body)
	})

	t.Run("should write no matched response when no stub match by method", func(t *testing.T) {
		httpReq, _ := http.NewRequest(http.MethodPost, "/api/users", http.NoBody)

		httpResp, err := client.Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})

	t.Run("should write no matched response when no stub match by url", func(t *testing.T) {
		httpReq, _ := http.NewRequest(http.MethodGet, "/api/users/john-doe", http.NoBody)

		httpResp, err := client.Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})
}

func TestWithSlogLogger(t *testing.T) {
	t.Parallel()

	t.Run("should log with the specified slog logger", func(t *testing.T) {
		var buff bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buff, &slog.HandlerOptions{Level: slog.LevelDebug}))

		server := mockaso.NewServer(mockaso.WithSlogLogger(logger, slog.LevelInfo))

		server.Logger().Log("test message")
		assert.Regexp(t, `time=[^ ]+ level=INFO msg="test message"`, buff.String())
	})
}

func TestWithLogLogger(t *testing.T) {
	t.Parallel()

	t.Run("should log with the specified log logger", func(t *testing.T) {
		var buff bytes.Buffer
		logger := log.New(&buff, "", 0)

		server := mockaso.NewServer(mockaso.WithLogLogger(logger))

		server.Logger().Log("test message")
		assert.Equal(t, "test message\n", buff.String())
	})
}

func assertNotMatchedResponse(t *testing.T, httpReq *http.Request, httpResp *http.Response) bool {
	t.Helper()

	expectedBody := fmt.Sprintf("no stubs for %s %s", httpReq.Method, httpReq.URL)

	return assert.Equal(t, 666, httpResp.StatusCode) &&
		assertBodyString(t, expectedBody, httpResp)
}

func assertBodyString(t *testing.T, expected string, r *http.Response) bool {
	t.Helper()
	return assert.Equal(t, expected, readString(r.Body))
}

func readString(r io.Reader) string {
	data, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return string(data)
}
