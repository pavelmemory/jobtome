package webhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pavelmemory/jobtome/internal/logging"
	"github.com/pavelmemory/jobtome/internal/shorten"
)

func TestShortenHandler_Create(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		logger := logging.NewTestLogger()
		r := NewRouter(logger)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockShortenService := NewMockShortenService(ctrl)
		mockShortenService.EXPECT().Create(gomock.Any(), shorten.Entity{URL: "https://example.com"}).Return(int64(1), nil)

		shortenHandler := NewShortenHandler(mockShortenService)
		shortenHandler.Register(r)

		req := httptest.NewRequest(http.MethodPost, "http://localhost/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
		req.Header.Set("content-type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		require.Equal(t, http.StatusCreated, resp.Code)
		require.Equal(t, "/api/shorten/1", resp.Header().Get("location"))
	})
}

func TestShortenHandler_Get(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		logger := logging.NewTestLogger()
		r := NewRouter(logger)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockShortenService := NewMockShortenService(ctrl)
		mockShortenService.EXPECT().Get(gomock.Any(), int64(1)).Return(shorten.Entity{ID: 1, Hash: "1", URL: "https://example.com"}, nil)

		shortenHandler := NewShortenHandler(mockShortenService)
		shortenHandler.Register(r)

		req := httptest.NewRequest(http.MethodGet, "http://localhost/api/shorten/1", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, "application/json; charset=utf-8", resp.Header().Get("content-type"))
		require.JSONEq(t, `{"id":1, "hash":"1", "url":"https://example.com"}`, resp.Body.String())
	})
}

func TestShortenHandler_List(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		logger := logging.NewTestLogger()
		r := NewRouter(logger)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		existing := []shorten.Entity{
			{ID: 1, Hash: "1", URL: "https://example.com"},
			{ID: 2, Hash: "2", URL: "https://stub.com"},
		}
		mockShortenService := NewMockShortenService(ctrl)
		mockShortenService.EXPECT().List(gomock.Any(), shorten.Pager{Limit: 10, Offset: 1}).Return(existing, nil)

		shortenHandler := NewShortenHandler(mockShortenService)
		shortenHandler.Register(r)

		req := httptest.NewRequest(http.MethodGet, "http://localhost/api/shorten?limit=10&offset=1", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, "application/json; charset=utf-8", resp.Header().Get("content-type"))
		require.JSONEq(t, `[{"id":1, "hash":"1", "url":"https://example.com"}, {"id":2, "hash":"2", "url":"https://stub.com"}]`, resp.Body.String())
	})
}

func TestShortenHandler_Delete(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		logger := logging.NewTestLogger()
		r := NewRouter(logger)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockShortenService := NewMockShortenService(ctrl)
		mockShortenService.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)

		shortenHandler := NewShortenHandler(mockShortenService)
		shortenHandler.Register(r)

		req := httptest.NewRequest(http.MethodDelete, "http://localhost/api/shorten/1", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		require.Equal(t, http.StatusNoContent, resp.Code)
	})
}
