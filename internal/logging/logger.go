package logging

import (
	"context"
	"sync"
)

// Logger is an abstraction used for logging in the service.
type Logger interface {
	// WithString adds `val` of `string` type to the logging context and returns it to the caller.
	WithString(key, val string) Logger
	// WithInt64 adds `val` of `int64` type to the logging context and returns it to the caller.
	WithInt64(key string, val int64) Logger
	// WithInt adds `val` of `int` type to the logging context and returns it to the caller.
	WithInt(key string, val int) Logger
	// WithError adds `err` to the logging context and returns it to the caller.
	WithError(err error) Logger
	// Debug flushes logging context with "debug" severity level.
	Debug(msg string)
	// Info flushes logging context with "info" severity level.
	Info(msg string)
	// Error flushes logging context with "error" severity level.
	Error(msg string)
	// IsDebug reports if logging level severity is higher than 'debug' level.
	IsDebug() bool
}

type ctxKey struct{}

// FromContext extracts logger from the context.
// It will panic if context has no associated logger in it.
// It should be use in a pair with `ToContext` that injects logger into context.
func FromContext(ctx context.Context) Logger {
	return ctx.Value(ctxKey{}).(Logger)
}

// ToContext injects a logger into the context.
// It should be use in a pair with `FromContext` that extracts logger back.
func ToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

func NewTestLogger() *TestLogger {
	return &TestLogger{entries: []map[string]interface{}{{}}}
}

// TestLogger should be used only for testing purposes.
type TestLogger struct {
	mtx     sync.Mutex
	entries []map[string]interface{}
}

func (tl *TestLogger) with(key string, val interface{}) {
	tl.mtx.Lock()
	defer tl.mtx.Unlock()

	tl.entries[len(tl.entries)-1][key] = val
}

func (tl *TestLogger) WithString(key, val string) Logger {
	tl.with(key, val)
	return tl
}

func (tl *TestLogger) WithInt64(key string, val int64) Logger {
	tl.with(key, val)
	return tl
}

func (tl *TestLogger) WithInt(key string, val int) Logger {
	tl.with(key, val)
	return tl
}

func (tl *TestLogger) WithError(err error) Logger {
	tl.with("error", err)
	return tl
}

func (tl *TestLogger) level(lvl, msg string) {
	tl.with("level", lvl)
	tl.with("msg", msg)

	tl.mtx.Lock()
	defer tl.mtx.Unlock()

	tl.entries = append(tl.entries, map[string]interface{}{})

}

func (tl *TestLogger) Debug(msg string) {
	tl.level("debug", msg)
}

func (tl *TestLogger) Info(msg string) {
	tl.level("info", msg)
}

func (tl *TestLogger) Error(msg string) {
	tl.level("error", msg)
}

func (tl *TestLogger) Entries() []map[string]interface{} {
	tl.mtx.Lock()
	defer tl.mtx.Unlock()

	return tl.entries
}

func (tl *TestLogger) IsDebug() bool {
	return true
}
