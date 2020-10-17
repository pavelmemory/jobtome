package shorten

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pavelmemory/jobtome/internal/storage"
	"github.com/pavelmemory/jobtome/internal/storage/migrations"
)

func initDB(t *testing.T, filepath string) (*storage.SQLLite, func()) {
	t.Helper()

	require.NoError(t, os.RemoveAll(filepath))

	require.NoError(t, migrations.Up(filepath))

	instance, err := storage.NewSQLLite(filepath)
	require.NoError(t, err)

	cleanup := func() {
		instance.Close()
		require.NoError(t, os.RemoveAll(filepath))
	}

	return instance, cleanup
}
