package shorten

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pavelmemory/jobtome/internal"
	"github.com/pavelmemory/jobtome/internal/storage/shorten"

	"github.com/pavelmemory/jobtome/internal/storage"
)

func TestService_Create(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		t.Run("no url", func(t *testing.T) {
			srv := NewService(nil, nil)
			_, err := srv.Create(Context(), Entity{URL: ""})
			exp := ValidationError{Cause: internal.ErrBadInput, Details: map[string]interface{}{"url": "blank or empty"}}
			require.Equal(t, exp, err)
		})

		t.Run("hash is set", func(t *testing.T) {
			srv := NewService(nil, nil)
			_, err := srv.Create(Context(), Entity{URL: "http://example.com", Hash: "stub"})
			exp := ValidationError{Cause: internal.ErrBadInput, Details: map[string]interface{}{"hash": "not empty"}}
			require.Equal(t, exp, err)
		})
	})

	t.Run("new", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().ByHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(shorten.Entity{}, internal.ErrNotFound)
		mockStorage.EXPECT().
			Persist(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, run storage.Runner, short shorten.Entity) (int64, error) {
				require.Equal(t, "https://example.com", short.URL)
				require.NotEmpty(t, "1234567", short.Hash)
				return 1, nil
			})

		srv := NewService(testTransactioner{}, mockStorage)
		id, err := srv.Create(Context(), Entity{URL: "https://example.com"})

		require.NoError(t, err)
		require.Equal(t, int64(1), id)
	})

	t.Run("reuse existing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		existing := shorten.Entity{ID: 1, URL: "https://example.com", Hash: "1234567"}
		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().ByHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(existing, nil)

		srv := NewService(testTransactioner{}, mockStorage)
		id, err := srv.Create(Context(), Entity{URL: "https://example.com"})

		require.NoError(t, err)
		require.Equal(t, existing.ID, id)
	})
}

func TestService_Get(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		const id = int64(1)

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().Retrieve(gomock.Any(), gomock.Any(), id).Return(shorten.Entity{}, internal.ErrNotFound)

		srv := NewService(testTransactioner{}, mockStorage)
		_, err := srv.Get(Context(), id)
		require.True(t, errors.Is(err, internal.ErrNotFound))
	})

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		existing := shorten.Entity{ID: 1, URL: "https://example.com", Hash: "1234567"}
		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().Retrieve(gomock.Any(), gomock.Any(), existing.ID).Return(existing, nil)

		srv := NewService(testTransactioner{}, mockStorage)
		actual, err := srv.Get(Context(), existing.ID)
		require.NoError(t, err)
		require.Equal(t, Entity{ID: existing.ID, URL: existing.URL, Hash: existing.Hash}, actual)
	})
}

func TestService_List(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		t.Run("bad limit", func(t *testing.T) {
			srv := NewService(nil, nil)
			_, err := srv.List(Context(), Pager{Limit: -50, Offset: 10})
			exp := ValidationError{Cause: internal.ErrBadInput, Details: map[string]interface{}{"limit": "is lesser then 1"}}
			require.Equal(t, exp, err)
		})

		t.Run("bad offset", func(t *testing.T) {
			srv := NewService(nil, nil)
			_, err := srv.List(Context(), Pager{Limit: 50, Offset: -10})
			exp := ValidationError{Cause: internal.ErrBadInput, Details: map[string]interface{}{"offset": "is negative"}}
			require.Equal(t, exp, err)
		})
	})

	t.Run("nothing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().List(gomock.Any(), gomock.Any(), shorten.Pager{Limit: 50, Offset: 10}).Return(nil, nil)

		srv := NewService(testTransactioner{}, mockStorage)
		actual, err := srv.List(Context(), Pager{Limit: 50, Offset: 10})
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		existing := []shorten.Entity{
			{ID: 1, URL: "https://example.com", Hash: "1"},
			{ID: 2, URL: "https://stub.com", Hash: "2"},
		}
		exp := []Entity{
			{ID: existing[0].ID, URL: existing[0].URL, Hash: existing[0].Hash},
			{ID: existing[1].ID, URL: existing[1].URL, Hash: existing[1].Hash},
		}
		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(existing, nil)

		srv := NewService(testTransactioner{}, mockStorage)
		actual, err := srv.List(Context(), Pager{Limit: 10})
		require.NoError(t, err)
		require.Equal(t, exp, actual)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		const id = int64(1)

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().Delete(gomock.Any(), gomock.Any(), id).Return(internal.ErrNotFound)

		srv := NewService(testTransactioner{}, mockStorage)
		err := srv.Delete(Context(), id)
		require.True(t, errors.Is(err, internal.ErrNotFound))
	})

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		srv := NewService(testTransactioner{}, mockStorage)
		err := srv.Delete(Context(), 1)
		require.NoError(t, err)
	})
}

func TestService_Resolve(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		t.Run("empty hash", func(t *testing.T) {
			srv := NewService(nil, nil)
			_, err := srv.Resolve(Context(), "   ")
			exp := ValidationError{Cause: internal.ErrBadInput, Details: map[string]interface{}{"hash": "blank or empty"}}
			require.Equal(t, exp, err)
		})
	})

	t.Run("not existing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		const hash = "1"

		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().ByHash(gomock.Any(), gomock.Any(), hash).Return(shorten.Entity{}, internal.ErrNotFound)

		srv := NewService(testTransactioner{}, mockStorage)
		_, err := srv.Resolve(Context(), hash)
		require.True(t, errors.Is(err, internal.ErrNotFound))
	})

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		existing := shorten.Entity{ID: 1, URL: "https://example.com", Hash: "1234567"}
		mockStorage := NewMockStorage(ctrl)
		mockStorage.EXPECT().ByHash(gomock.Any(), gomock.Any(), existing.Hash).Return(existing, nil)

		srv := NewService(testTransactioner{}, mockStorage)
		actual, err := srv.Resolve(Context(), existing.Hash)
		require.NoError(t, err)
		require.Equal(t, existing.URL, actual)
	})
}

// TODO: verify flow when Transactioner fails to commit

func Context() context.Context {
	return context.Background()
}

type testTransactioner struct{}

func (testTransactioner) WithTx(_ context.Context, call func(runner storage.Runner) error) error {
	return call(nil)
}

func (testTransactioner) WithoutTx(_ context.Context, call func(runner storage.Runner) error) error {
	return call(nil)
}
