package migrations

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Up(filepath string) error {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}

	for i, migration := range []func(*sql.DB) error{
		UrlShortened,
	} {
		if err := migration(db); err != nil {
			return fmt.Errorf("apply migration %d: %w", i+1, err)
		}
	}

	return nil
}
