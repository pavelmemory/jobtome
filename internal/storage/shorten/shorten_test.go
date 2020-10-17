package shorten

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pavelmemory/jobtome/internal/storage"

	"github.com/pavelmemory/jobtome/internal"
)

func TestSQLLite_Persist(t *testing.T) {
	db, cleanup := initDB(t, t.Name())
	defer cleanup()

	repo := Repo{}

	t.Run("bad input", func(t *testing.T) {
		t.Run("url", func(t *testing.T) {
			err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
				_, err := repo.Persist(context.Background(), runner, Entity{URL: "", Hash: "hash"})
				require.Error(t, err)
				require.True(t, errors.Is(err, internal.ErrBadInput), err.Error())
				require.Contains(t, err.Error(), "CHECK constraint failed")
				return nil
			})
			require.NoError(t, err)
		})

		t.Run("hash", func(t *testing.T) {
			err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
				_, err := repo.Persist(context.Background(), runner, Entity{URL: "url", Hash: ""})
				require.Error(t, err)
				require.True(t, errors.Is(err, internal.ErrBadInput), err.Error())
				require.Contains(t, err.Error(), "CHECK constraint failed")
				return nil
			})
			require.NoError(t, err)
		})
	})

	t.Run("ok", func(t *testing.T) {
		var id int64

		err := db.WithoutTx(context.Background(), func(runner storage.Runner) (err error) {
			id, err = repo.Persist(context.Background(), runner, Entity{URL: "https://example.com", Hash: "12345"})
			return err
		})
		require.NoError(t, err)

		require.GreaterOrEqual(t, int64(1), id)

		err = db.WithTx(context.Background(), func(runner storage.Runner) error {
			var url, hash string
			var createdAt int64
			res := runner.QuerySingle(context.Background(), "SELECT url, hash, created_at FROM shorten WHERE id = $1", id)
			require.NoError(t, res.Scan(&url, &hash, &createdAt))
			require.Equal(t, "12345", hash)
			require.Equal(t, "https://example.com", url)
			require.Greater(t, createdAt, time.Now().Add(-time.Minute).Unix())
			require.Less(t, createdAt, time.Now().Add(time.Minute).Unix())
			return nil
		})
		require.NoError(t, err)
	})
}

func TestSQLLite_Retrieve(t *testing.T) {
	db, cleanup := initDB(t, t.Name())
	defer cleanup()

	repo := Repo{}

	t.Run("nothing", func(t *testing.T) {
		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			_, err := repo.Retrieve(context.Background(), runner, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, internal.ErrNotFound), err.Error())
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		now := time.Now()
		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			insert(t, runner, Entity{URL: "1", Hash: "https://example.com", CreatedAt: now})
			return nil
		})
		require.NoError(t, err)

		err = db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			entity, err := repo.Retrieve(context.Background(), runner, 1)
			require.NoError(t, err)
			require.Equal(t, int64(1), entity.ID)
			require.Equal(t, "1", entity.URL)
			require.Equal(t, "https://example.com", entity.Hash)
			require.Equal(t, now.Unix(), entity.CreatedAt.Unix())
			return nil
		})
		require.NoError(t, err)
	})
}

func TestSQLLite_List(t *testing.T) {
	db, cleanup := initDB(t, t.Name())
	defer cleanup()

	repo := Repo{}

	t.Run("nothing", func(t *testing.T) {
		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			entities, err := repo.List(context.Background(), runner, Pager{Limit: 100})
			require.NoError(t, err)
			require.Nil(t, entities)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		now := time.Now()

		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			insert(t, runner, Entity{Hash: "1", URL: "https://example.com", CreatedAt: now})
			insert(t, runner, Entity{Hash: "2", URL: "https://stub.com", CreatedAt: now})
			return nil
		})
		require.NoError(t, err)

		t.Run("limited", func(t *testing.T) {
			err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
				entities, err := repo.List(context.Background(), runner, Pager{Limit: 1, Offset: 1})
				require.NoError(t, err)
				require.Len(t, entities, 1)
				require.Equal(t, "2", entities[0].Hash)
				require.Equal(t, "https://stub.com", entities[0].URL)
				require.Equal(t, now.Unix(), entities[0].CreatedAt.Unix())
				return nil
			})
			require.NoError(t, err)
		})

		t.Run("all", func(t *testing.T) {
			err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
				entities, err := repo.List(context.Background(), runner, Pager{Limit: 100})
				require.NoError(t, err)
				require.Len(t, entities, 2)
				require.Equal(t, "1", entities[0].Hash)
				require.Equal(t, "2", entities[1].Hash)
				return nil
			})
			require.NoError(t, err)
		})
	})
}

func TestSQLLite_Delete(t *testing.T) {
	db, cleanup := initDB(t, t.Name())
	defer cleanup()

	repo := Repo{}

	t.Run("not existing", func(t *testing.T) {
		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			err := repo.Delete(context.Background(), runner, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, internal.ErrNotFound), err.Error())
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		now := time.Now()
		var id int64
		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			id = insert(t, runner, Entity{Hash: "1", URL: "https://example.com", CreatedAt: now})
			return nil
		})
		require.NoError(t, err)

		err = db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			err := repo.Delete(context.Background(), runner, id)
			require.NoError(t, err)

			var dst interface{}
			res := runner.QuerySingle(context.Background(), "SELECT 1 FROM shorten WHERE id = $1", id)
			err = res.Scan(&dst)
			require.Error(t, err)
			require.True(t, errors.Is(err, sql.ErrNoRows), err.Error())
			return nil
		})
		require.NoError(t, err)
	})
}

func TestSQLLite_ByHash(t *testing.T) {
	db, cleanup := initDB(t, t.Name())
	defer cleanup()

	repo := Repo{}

	t.Run("nothing", func(t *testing.T) {
		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			_, err := repo.ByHash(context.Background(), runner, "1234567")
			require.Error(t, err)
			require.True(t, errors.Is(err, internal.ErrNotFound), err.Error())
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		now := time.Now()
		err := db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			insert(t, runner, Entity{URL: "https://example.com", Hash: "1234567", CreatedAt: now})
			return nil
		})
		require.NoError(t, err)

		err = db.WithoutTx(context.Background(), func(runner storage.Runner) error {
			entity, err := repo.ByHash(context.Background(), runner, "1234567")
			require.NoError(t, err)
			require.Equal(t, int64(1), entity.ID)
			require.Equal(t, "https://example.com", entity.URL)
			require.Equal(t, "1234567", entity.Hash)
			require.Equal(t, now.Unix(), entity.CreatedAt.Unix())
			return nil
		})
		require.NoError(t, err)
	})
}

func insert(t *testing.T, runner storage.Runner, shorten Entity) int64 {
	res := runner.Exec(
		context.Background(),
		`INSERT INTO shorten VALUES (NULL, $1, $2, $3)`,
		shorten.URL, shorten.Hash, shorten.CreatedAt.Unix(),
	)
	require.NoError(t, res.Err())
	require.EqualValues(t, 1, res.Affected())
	return res.ID()
}
