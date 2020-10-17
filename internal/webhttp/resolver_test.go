package webhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pavelmemory/jobtome/internal/logging"
)

func TestResolverHandler_Resolve(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		logger := logging.NewTestLogger()
		r := NewRouter(logger)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockShortenService := NewMockShortenService(ctrl)
		mockShortenService.EXPECT().Resolve(gomock.Any(), "hash").Return("https://example.com", nil)

		resolverHandler := NewResolverHandler(mockShortenService)
		resolverHandler.Register(r)

		req := httptest.NewRequest(http.MethodGet, "http://localhost/hash", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		require.Equal(t, http.StatusTemporaryRedirect, resp.Code)
		require.Equal(t, "https://example.com", resp.Header().Get("location"))
	})
}
