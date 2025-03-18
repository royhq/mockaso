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

	server.Stub(http.MethodGet, mockaso.Path("/test/match")).
		Match(
			mockaso.MatchRequest(matchOnlyJohn),
		).
		Respond(
			mockaso.WithStatusCode(http.StatusOK),
			mockaso.WithBody("matched request"),
		)

	t.Run("should return the specified stub when request match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, "/test/match?name=john", http.NoBody)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assertBodyString(t, "matched request", httpResp)
	})

	t.Run("should return no match response when request does not match", func(t *testing.T) {
		t.Parallel()

		httpReq, _ := http.NewRequest(http.MethodGet, "/test/match?name=rick", http.NoBody)
		httpResp, err := server.Client().Do(httpReq)
		require.NoError(t, err)

		assertNotMatchedResponse(t, httpReq, httpResp)
	})
}
