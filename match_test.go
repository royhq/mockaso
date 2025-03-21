package mockaso_test

import (
	"net/http"
	"net/http/httptest"
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
		Match(
			mockaso.MatchRequest(matchOnlyJohn),
		).
		Respond(
			mockaso.WithStatusCode(http.StatusOK),
			mockaso.WithBody("matched request"),
		)

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
		Match(
			mockaso.MatchHeader("X-Test-Header", "test value"),
		).
		Respond(
			mockaso.WithStatusCode(http.StatusOK),
			mockaso.WithBody("matched request"),
		)

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
