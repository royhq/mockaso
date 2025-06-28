package mockaso_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

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
			body         any
			expectedBody string
		}{
			"bytes body": {
				body:         []byte("test bytes body"),
				expectedBody: `test bytes body`,
			},
			"string body": {
				body:         "test string body",
				expectedBody: `test string body`,
			},
			"int body": {
				body:         123,
				expectedBody: `123`,
			},
			"json raw body": {
				body:         json.RawMessage(`{"name":"john"}`),
				expectedBody: `{"name":"john"}`,
			},
			"string reader body": {
				body:         strings.NewReader("string reader body"),
				expectedBody: `string reader body`,
			},
			"buffer body": {
				body:         bytes.NewBuffer([]byte("buffer body")),
				expectedBody: `buffer body`,
			},
			"map body": {
				body:         map[string]any{"name": "john", "age": 57},
				expectedBody: `map[age:57 name:john]`,
			},
			"struct body": {
				body:         userResponse{Name: "john", Age: 57},
				expectedBody: `{john 57}`,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				url := fmt.Sprintf("/test/%s", strings.ReplaceAll(name, " ", "-"))
				server.Stub(http.MethodGet, mockaso.URL(url)).
					Respond(
						mockaso.WithStatusCode(http.StatusOK),
						mockaso.WithBody(tc.body),
					)

				httpReq, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
				httpResp, err := server.Client().Do(httpReq)
				require.NoError(t, err)

				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				assertBodyString(t, tc.expectedBody, httpResp)
			})
		}
	})
}

func TestWithRawJSON(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	t.Run("should return the specified json", func(t *testing.T) {
		testCases := map[string]struct {
			rule         mockaso.StubResponseRule
			expectedBody string
		}{
			"json raw object as string": {
				rule:         mockaso.WithRawJSON(json.RawMessage(`{"name":"john","age":57}`)),
				expectedBody: `{"name":"john","age":57}`,
			},
			"object as string": {
				rule:         mockaso.WithRawJSON(`{"name":"rick","age":39}`),
				expectedBody: `{"name":"rick","age":39}`,
			},
			"object as bytes": {
				rule:         mockaso.WithRawJSON(`{"name":"carl","age":21}`),
				expectedBody: `{"name":"carl","age":21}`,
			},
			"raw string": {
				rule:         mockaso.WithRawJSON(json.RawMessage(`"john"`)),
				expectedBody: `"john"`,
			},
			"string": {
				rule:         mockaso.WithRawJSON(`"rick"`),
				expectedBody: `"rick"`,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				url := fmt.Sprintf("/test/with-raw-json/%s", strings.ReplaceAll(name, " ", "-"))
				server.Stub(http.MethodGet, mockaso.URL(url)).
					Respond(
						mockaso.WithStatusCode(http.StatusOK),
						tc.rule,
					)

				httpReq, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
				httpResp, err := server.Client().Do(httpReq)
				require.NoError(t, err)

				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				assert.Equal(t, "application/json", httpResp.Header.Get("Content-Type"))
				assertBodyString(t, tc.expectedBody, httpResp)
			})
		}
	})

	t.Run("should panic when json is not valid", func(t *testing.T) {
		t.Parallel()

		testCases := map[string]struct {
			body string
		}{
			"invalid object": {
				body: `{"name":"john",}`,
			},
			"invalid string": {
				body: `john`,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				fn := func() {
					url := fmt.Sprintf("/test/with-raw-json/panic/%s", strings.ReplaceAll(name, " ", "-"))
					server.Stub(http.MethodGet, mockaso.URL(url)).
						Respond(mockaso.WithRawJSON(tc.body))
				}

				assert.Panics(t, fn)
			})
		}
	})
}

func TestWithJSON(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	t.Run("should return the specified json", func(t *testing.T) {
		testCases := map[string]struct {
			body         any
			expectedBody string
		}{
			"int": {
				body:         123,
				expectedBody: `123`,
			},
			"float": {
				body:         20.87,
				expectedBody: `20.87`,
			},
			"string": {
				body:         `john`,
				expectedBody: `"john"`,
			},
			"map": {
				body:         map[string]any{"name": "john", "age": 57},
				expectedBody: `{"age":57,"name":"john"}`,
			},
			"struct": {
				body:         userResponse{Name: "rick", Age: 39},
				expectedBody: `{"name":"rick","age":39}`,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				url := fmt.Sprintf("/test/with-json/%s", strings.ReplaceAll(name, " ", "-"))
				server.Stub(http.MethodGet, mockaso.URL(url)).
					Respond(
						mockaso.WithStatusCode(http.StatusOK),
						mockaso.WithJSON(tc.body),
					)

				httpReq, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
				httpResp, err := server.Client().Do(httpReq)
				require.NoError(t, err)

				assert.Equal(t, http.StatusOK, httpResp.StatusCode)
				assert.Equal(t, "application/json", httpResp.Header.Get("Content-Type"))
				assertBodyString(t, tc.expectedBody, httpResp)
			})
		}
	})

	t.Run("should panic when the json is not valid", func(t *testing.T) {
		t.Parallel()

		fn := func() {
			server.Stub(http.MethodGet, mockaso.URL("/test/with-json/panic")).
				Respond(mockaso.WithJSON(invalidJSON("any")))
		}

		assert.Panics(t, fn)
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

func TestWithDelay(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	t.Run("should return with the specified delay", func(t *testing.T) {
		url := "/test/with-delay"
		delay := 1200 * time.Millisecond
		start := time.Now()

		server.Stub(http.MethodGet, mockaso.URL(url)).
			Respond(
				mockaso.WithStatusCode(http.StatusOK),
				mockaso.WithDelay(delay),
			)

		httpReq, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
		httpResp, err := server.Client().Do(httpReq)
		elapsed := time.Since(start)

		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		t.Logf("duration: %v", elapsed)
		assert.GreaterOrEqual(t, elapsed, delay)
	})
}

type userResponse struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type invalidJSON string

func (p invalidJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("invalid json")
}
