package mockaso_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

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
