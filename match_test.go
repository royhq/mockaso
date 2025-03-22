package mockaso_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/royhq/mockaso"
)

func TestURL(t *testing.T) {
	t.Parallel()

	reqURL := "/api/users?page=1&size=20"
	httpReq := httptest.NewRequest(http.MethodGet, reqURL, http.NoBody)

	testCases := map[string]struct {
		matchURL      string
		expectedMatch bool
	}{
		"should return true when url matched": {
			matchURL:      "/api/users?page=1&size=20",
			expectedMatch: true,
		},
		"should return false when url does not match (query param diff)": {
			matchURL:      "/api/users?page=2&size=20",
			expectedMatch: false,
		},
		"should return false when url does not match (missing query param)": {
			matchURL:      "/api/users?page=1",
			expectedMatch: false,
		},
		"should return false when url does not match (missing all query params)": {
			matchURL:      "/api/users",
			expectedMatch: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			matcher := mockaso.URL(tc.matchURL)
			assert.Equal(t, tc.expectedMatch, matcher(httpReq.URL))
		})
	}
}

func TestPath(t *testing.T) {
	t.Parallel()

	reqURL := "/api/users?page=1&size=20"
	httpReq := httptest.NewRequest(http.MethodGet, reqURL, http.NoBody)

	testCases := map[string]struct {
		matchURL      string
		expectedMatch bool
	}{
		"should return true when path matched": {
			matchURL:      "/api/users",
			expectedMatch: true,
		},
		"should return true when path matched with trailing slash": {
			matchURL:      "/api/users/",
			expectedMatch: true,
		},
		"should return false when path does not match": {
			matchURL:      "/api/users/john-doe",
			expectedMatch: false,
		},
		"should return false when path does not match by including query params": {
			matchURL:      "/api/users?page=1&size=20",
			expectedMatch: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			matcher := mockaso.Path(tc.matchURL)
			assert.Equal(t, tc.expectedMatch, matcher(httpReq.URL))
		})
	}
}

func TestURLRegex(t *testing.T) {
	t.Parallel()

	reqURL := "/api/users?page=1&size=20"
	httpReq := httptest.NewRequest(http.MethodGet, reqURL, http.NoBody)

	regexValues := []string{
		`\/api\/users\?page=1\&size=20`,
		`^\/api\/users\?page=1\&size=20$`,
		`^\/api\/users\?page=\d+\&size=\d+$`,
		`^\/api\/[a-zA-Z]+\?page=\d+\&size=\d+$`,
	}

	for _, r := range regexValues {
		t.Run(r, func(t *testing.T) {
			t.Parallel()
			matcher := mockaso.URLRegex(r)
			assert.True(t, matcher(httpReq.URL))
		})
	}
}

func TestPathRegex(t *testing.T) {
	t.Parallel()

	reqURL := "/api/users?page=1&size=20"
	httpReq := httptest.NewRequest(http.MethodGet, reqURL, http.NoBody)

	regexValues := []string{
		`\/api\/users`,
		`^\/api\/users$`,
		`^\/api\/[a-zA-Z]+`,
	}

	for _, r := range regexValues {
		t.Run(r, func(t *testing.T) {
			t.Parallel()
			matcher := mockaso.PathRegex(r)
			assert.True(t, matcher(httpReq.URL))
		})
	}
}

func TestMatchRequest(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	var calls atomic.Int32
	matchOnlyJohn := mockaso.RequestMatcherFunc(func(r *http.Request) bool {
		calls.Add(1)
		return r.URL.Query().Get("name") == "john"
	})

	t.Cleanup(func() {
		assert.Equal(t, int32(2), calls.Load())
	})

	const path = "/test/match-request"

	server.Stub(http.MethodGet, mockaso.Path(path)).
		Match(mockaso.MatchRequest(matchOnlyJohn)).
		Respond(matchedRequestRules()...)

	t.Run("should return the specified stub when request match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, path+"?name=john", http.NoBody)
		require.Equal(t, path, httpReq.URL.Path)

		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when request does not match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, path+"?name=rick", http.NoBody)
		require.Equal(t, path, httpReq.URL.Path)

		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})
}

func TestMatchHeader(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	const path = "/test/match-header"

	server.Stub(http.MethodGet, mockaso.Path(path)).
		Match(mockaso.MatchHeader("X-Test-Header", "test value")).
		Respond(matchedRequestRules()...)

	t.Run("should return the specified stub when header match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, path, http.NoBody)
		httpReq.Header.Set("X-Test-Header", "test value")

		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when header does not match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, path, http.NoBody)
		httpReq.Header.Set("X-Test-Header", "another test value")

		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})
}

func TestMatchQuery(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	const path = "/test/match-query"

	server.Stub(http.MethodGet, mockaso.Path(path)).
		Match(mockaso.MatchQuery("name", "john")).
		Respond(matchedRequestRules()...)

	t.Run("should return the specified stub when query match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, path+"?name=john", http.NoBody)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when query does not match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, path+"?name=rick", http.NoBody)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})
}

func TestMatchNoBody(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	const path = "/test/match-no-body"

	server.Stub(http.MethodGet, mockaso.Path(path)).
		Match(mockaso.MatchNoBody()).
		Respond(matchedRequestRules()...)

	t.Run("should return the specified stub when request has no body", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, path, http.NoBody)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when request has body", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`request body`)
		httpReq, _ := http.NewRequest(http.MethodGet, path, body)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})
}

func TestMatchRawJSONBody(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	const path = "/test/match-raw-json"

	server.Stub(http.MethodPost, mockaso.Path(path)).
		Match(mockaso.MatchRawJSONBody(`{"name":"john"}`)).
		Respond(matchedRequestRules()...)

	t.Run("should return the specified stub when request match", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"john"}`)
		httpReq, _ := http.NewRequest(http.MethodPost, path, body)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when request does not match", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"rick"}`)
		httpReq, _ := http.NewRequest(http.MethodPost, path, body)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})
}

func TestMatchJSONBody(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	const path = "/test/match-json-body"

	t.Run("should return the specified stub", func(t *testing.T) {
		t.Run("when specified body is a map", func(t *testing.T) {
			t.Parallel()

			server.Stub(http.MethodPost, mockaso.Path(path+"/map")).
				Match(mockaso.MatchJSONBody(map[string]string{"name": "john"})).
				Respond(matchedRequestRules()...)

			body := strings.NewReader(`{"name":"john"}`)
			httpReq, _ := http.NewRequest(http.MethodPost, path+"/map", body)
			httpResp, err := server.Client().Do(httpReq)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, httpResp.StatusCode)
			assertBodyString(t, "matched request", httpResp)
		})
	})
}

func TestMatchBodyMapFunc(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	var calls atomic.Int32
	matchOnlyJohn := mockaso.BodyMatcherMapFunc(func(body map[string]any) bool {
		calls.Add(1)
		return body["name"] == "john"
	})

	t.Cleanup(func() {
		assert.Equal(t, int32(2), calls.Load())
	})

	const path = "/test/body-as-map"

	server.Stub(http.MethodPost, mockaso.Path(path)).
		Match(mockaso.MatchBodyMapFunc(matchOnlyJohn)).
		Respond(matchedRequestRules()...)

	t.Run("should return the specified stub when matcher is true", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"john"}`)
		httpReq, _ := http.NewRequest(http.MethodPost, path, body)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when matcher is false", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"rick"}`)
		httpReq, _ := http.NewRequest(http.MethodPost, path, body)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})

	t.Run("should receive an empty map in matcher when request has no body", func(t *testing.T) {
		t.Parallel()

		const path = path + "/empty-body"

		matcher := mockaso.BodyMatcherMapFunc(func(body map[string]any) bool {
			assert.NotNil(t, body)
			assert.Empty(t, body)

			return true
		})

		server.Stub(http.MethodPost, mockaso.Path(path)).
			Match(mockaso.MatchBodyMapFunc(matcher)).
			Respond(matchedRequestRules()...)

		httpReq, _ := http.NewRequest(http.MethodPost, path, http.NoBody)
		require.Equal(t, path, httpReq.URL.Path)

		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})
}

func TestMatchBodyStringFunc(t *testing.T) {
	t.Parallel()

	server := mockaso.MustStartNewServer(mockaso.WithLogger(t))
	t.Cleanup(server.MustShutdown)

	var calls atomic.Int32
	matchOnlyJohn := mockaso.BodyMatcherStringFunc(func(body string) bool {
		calls.Add(1)
		return strings.Contains(body, `:"john"`)
	})

	t.Cleanup(func() {
		assert.Equal(t, int32(2), calls.Load())
	})

	const path = "/test/body-as-string"

	server.Stub(http.MethodPost, mockaso.Path(path)).
		Match(mockaso.MatchBodyStringFunc(matchOnlyJohn)).
		Respond(matchedRequestRules()...)

	t.Run("should return the specified stub when matcher is true", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"john"}`)
		httpReq, _ := http.NewRequest(http.MethodPost, path, body)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when matcher is false", func(t *testing.T) {
		t.Parallel()

		body := strings.NewReader(`{"name":"rick"}`)
		httpReq, _ := http.NewRequest(http.MethodPost, path, body)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})

	t.Run("should receive an empty string in matcher when request has no body", func(t *testing.T) {
		t.Parallel()

		const path = path + "/empty-body"

		matcher := mockaso.BodyMatcherStringFunc(func(body string) bool {
			assert.Empty(t, body)
			return true
		})

		server.Stub(http.MethodPost, mockaso.Path(path)).
			Match(mockaso.MatchBodyStringFunc(matcher)).
			Respond(matchedRequestRules()...)

		httpReq, _ := http.NewRequest(http.MethodPost, path, http.NoBody)
		require.Equal(t, path, httpReq.URL.Path)

		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})
}

func matchedRequestRules() []mockaso.StubResponseRule {
	return []mockaso.StubResponseRule{
		mockaso.WithStatusCode(http.StatusOK),
		mockaso.WithBody("matched request"),
	}
}
