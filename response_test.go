package mockaso_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/royhq/mockaso"
)

func TestWithStatusCode(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	t.Run("should return the specified status code", func(t *testing.T) {
		statusCodes := []int{http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusBadRequest,
			http.StatusNotFound, http.StatusInternalServerError, http.StatusServiceUnavailable}

		for _, statusCode := range statusCodes {
			t.Run(fmt.Sprintf("with status code %d", statusCode), func(t *testing.T) {
				t.Parallel()

				url := fmt.Sprintf("/test/%d", statusCode)
				server.Stub(http.MethodGet, mockaso.URL(url)).Respond(mockaso.WithStatusCode(statusCode))

				httpReq, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
				httpResp, err := server.Client().Do(httpReq)
				require.NoError(t, err)

				assert.Equal(t, statusCode, httpResp.StatusCode)
			})
		}
	})

	t.Run("should return http 200 when status code is not specified", func(t *testing.T) {
		t.Parallel()

		server.Stub(http.MethodGet, mockaso.URL("/test"))

		httpReq, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	})
}

func TestWithBody(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	t.Run("should return the specified body", func(t *testing.T) {
		testCases := map[string]struct {
			url          string
			body         any
			expectedBody string
		}{
			"bytes body": {
				url:          "/test/bytes",
				body:         []byte("test bytes body"),
				expectedBody: `test bytes body`,
			},
			"string body": {
				url:          "/test/string",
				body:         "test string body",
				expectedBody: `test string body`,
			},
			"int body": {
				url:          "/test/int",
				body:         123,
				expectedBody: `123`,
			},
			"json raw body": {
				url:          "/test/json",
				body:         json.RawMessage(`{"name":"john"}`),
				expectedBody: `{"name":"john"}`,
			},
			"string reader body": {
				url:          "/test/string-reader",
				body:         strings.NewReader("string reader body"),
				expectedBody: `string reader body`,
			},
			"buffer body": {
				url:          "/test/buffer",
				body:         bytes.NewBuffer([]byte("buffer body")),
				expectedBody: `buffer body`,
			},
			"map body": {
				url:          "/test/map",
				body:         map[string]any{"name": "john", "age": 57},
				expectedBody: `map[age:57 name:john]`,
			},
			"struct body": {
				url:          "/test/struct",
				body:         userResponse{Name: "john", Age: 57},
				expectedBody: `{john 57}`,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				server.Stub(http.MethodGet, mockaso.URL(tc.url)).
					Respond(
						mockaso.WithStatusCode(http.StatusOK),
						mockaso.WithBody(tc.body),
					)

				httpReq, _ := http.NewRequest(http.MethodGet, tc.url, http.NoBody)
				httpResp, err := server.Client().Do(httpReq)
				require.NoError(t, err)

				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				assertBodyString(t, tc.expectedBody, httpResp)
			})
		}
	})
}

func TestWithHeader_And_WithHeaders(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	t.Run("WithHeader", func(t *testing.T) {
		t.Run("should return the specified header", func(t *testing.T) {
			url := "/test/with-header"

			server.Stub(http.MethodGet, mockaso.URL(url)).
				Respond(
					mockaso.WithStatusCode(http.StatusOK),
					mockaso.WithHeader("X-Test-Header1", "test value 1"),
					mockaso.WithHeader("X-Test-Header2", "test value 2a"),
					mockaso.WithHeader("X-Test-Header2", "test value 2b"),
				)

			httpReq, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
			httpResp, err := server.Client().Do(httpReq)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, httpResp.StatusCode)
			assert.Equal(t, "test value 1", httpResp.Header.Get("X-Test-Header1"))
			assert.Equal(t, "test value 2b", httpResp.Header.Get("X-Test-Header2"))
		})
	})

	t.Run("WithHeaders", func(t *testing.T) {
		t.Run("should return the specified headers", func(t *testing.T) {
			url := "/test/with-headers"

			server.Stub(http.MethodGet, mockaso.URL(url)).
				Respond(
					mockaso.WithStatusCode(http.StatusOK),
					mockaso.WithHeader("X-Test-Header1", "test value 1"),
					mockaso.WithHeaders(map[string]string{
						"X-Test-Header2": "test value 2a",
						"X-Test-Header3": "test value 3",
					}),
					mockaso.WithHeaders(map[string]string{
						"X-Test-Header2": "test value 2b",
						"X-Test-Header4": "test value 4",
					}),
				)

			httpReq, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
			httpResp, err := server.Client().Do(httpReq)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, httpResp.StatusCode)
			assert.Equal(t, "test value 1", httpResp.Header.Get("X-Test-Header1"))
			assert.Equal(t, "test value 2b", httpResp.Header.Get("X-Test-Header2"))
			assert.Equal(t, "test value 3", httpResp.Header.Get("X-Test-Header3"))
			assert.Equal(t, "test value 4", httpResp.Header.Get("X-Test-Header4"))
		})
	})
}

type userResponse struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
