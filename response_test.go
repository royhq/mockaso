package mockaso_test

import (
	"fmt"
	"net/http"
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
