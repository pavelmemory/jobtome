package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"

	"github.com/pavelmemory/jobtome/internal"
)

type ExecResult interface {
	Err() error
	Affected() int64
	ID() int64
}

type MultiResult interface {
	Next() bool
	Scan(dst ...interface{}) error
	Close() error
}

type SingleResult interface {
	Scan(dst ...interface{}) error
}

type Runner interface {
	Exec(ctx context.Context, query string, params ...interface{}) ExecResult
	Query(ctx context.Context, query string, params ...interface{}) (MultiResult, error)
	QuerySingle(ctx context.Context, query string, params ...interface{}) SingleResult
}

// NewSQLLite returns a connection pool ready to execute statements on SQLLite database.
func NewSQLLite(filepath string) (*SQLLite, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping databse: %w", err)
	}

	return &SQLLite{db: db}, nil
}

type SQLLite struct {
	db *sql.DB
}

func (p *SQLLite) WithTx(ctx context.Context, action func(runner Runner) error) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := action(txRunner{tx: tx}); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (p *SQLLite) WithoutTx(_ context.Context, action func(runner Runner) error) error {
	return action(qRunner{db: p.db})
}

func (p *SQLLite) Close() {
	p.db.Close()
}

type qRunner struct {
	db *sql.DB
}

func (q qRunner) Exec(ctx context.Context, query string, params ...interface{}) ExecResult {
	res, err := q.db.ExecContext(ctx, query, params...)
	if err != nil {
		return execResult{err: err}
	}

	n, err := res.RowsAffected()
	if err != nil {
		return execResult{err: err}
	}

	id, err := res.LastInsertId()
	return execResult{result: n, id: id, err: err}
}

func (q qRunner) Query(ctx context.Context, query string, params ...interface{}) (MultiResult, error) {
	return q.db.QueryContext(ctx, query, params...)
}

func (q qRunner) QuerySingle(ctx context.Context, query string, params ...interface{}) SingleResult {
	return q.db.QueryRowContext(ctx, query, params...)
}

type txRunner struct {
	tx *sql.Tx
}

func (r txRunner) Exec(ctx context.Context, query string, params ...interface{}) ExecResult {
	res, err := r.tx.ExecContext(ctx, query, params...)
	if err != nil {
		return execResult{err: err}
	}

	n, err := res.RowsAffected()
	return execResult{result: n, err: err}
}

func (r txRunner) Query(ctx context.Context, query string, params ...interface{}) (MultiResult, error) {
	return r.tx.QueryContext(ctx, query, params...)
}

func (r txRunner) QuerySingle(ctx context.Context, query string, params ...interface{}) SingleResult {
	return r.tx.QueryRowContext(ctx, query, params...)
}

type execResult struct {
	result int64
	id     int64
	err    error
}

func (r execResult) Err() error {
	return r.err
}

func (r execResult) Affected() int64 {
	return r.result
}

func (r execResult) ID() int64 {
	return r.id
}

func ConvertError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return internal.ErrNotFound
	}

	var terr sqlite3.Error
	if !errors.As(err, &terr) {
		return err
	}

	var cause error
	switch terr.Code {
	case sqlite3.ErrConstraint, sqlite3.ErrTooBig, sqlite3.ErrMismatch:
		if terr.ExtendedCode == sqlite3.ErrConstraintUnique {
			cause = internal.ErrNotUnique
		} else {
			cause = internal.ErrBadInput
		}
	default:
		return err
	}

	return fmt.Errorf("%v: %w", err, cause)
}
