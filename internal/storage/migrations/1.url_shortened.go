package migrations

import (
	"database/sql"
)

func UrlShortened(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS shorten (
			id INTEGER PRIMARY KEY,
			url TEXT NOT NULL CHECK(LENGTH(url) > 0),
			hash TEXT NOT NULL CHECK(LENGTH(hash) > 0) UNIQUE,
			created_at INTEGER NOT NULL
		)`)
	if err != nil {
		return err
	}

	return tx.Commit()
}
