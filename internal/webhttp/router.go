package webhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/pavelmemory/jobtome/internal"
	"github.com/pavelmemory/jobtome/internal/logging"
)

// NewRouter returns initialized HTTP router.
// It sets up all required middlewares and bindings for endpoints.
func NewRouter(logger logging.Logger) chi.Router {
	router := chi.NewRouter()
	router.Use(InjectLogger(logger)) // TODO: CORS, caching, tracing, metrics, etc.

	router.With(LogRequest()).NotFound(undefined)
	router.With(LogRequest()).MethodNotAllowed(undefined)

	return router
}

func undefined(w http.ResponseWriter, r *http.Request) {
	logging.FromContext(r.Context()).Debug("request is not implemented")
	w.WriteHeader(http.StatusNotImplemented)
}

// WriteError sends an error response back to the client.
func WriteError(w http.ResponseWriter, logger logging.Logger, err error) {
	resp := ErrorResponse{StatusCode: http.StatusInternalServerError}
	if logger.IsDebug() {
		// sends error details back to the client only in debugging mode
		resp.Cause = err
	}

	switch {
	case errors.Is(err, internal.ErrBadInput):
		resp.StatusCode = http.StatusBadRequest
	case errors.Is(err, internal.ErrNotUnique):
		resp.StatusCode = http.StatusConflict
	case errors.Is(err, internal.ErrNotFound):
		resp.StatusCode = http.StatusNotFound
	}

	resp.Write(logger, w)
}

var (
	ErrMissingRequired = errors.New("missing required")
	ErrBadFormat       = errors.New("bad format")
)

// ErrorResponse aggregates error information into the struct and know how to send it back to the client.
type ErrorResponse struct {
	// StatusCode status code that should be returned.
	StatusCode int
	// Cause is an error that needs to be placed as a message describing what was wrong.
	Cause error
}

func (er ErrorResponse) Write(logger logging.Logger, w http.ResponseWriter) {
	w.WriteHeader(er.StatusCode)

	if er.Cause == nil {
		return
	}

	if err := Encode(w, er.Cause.Error()); err != nil {
		logger.WithError(err).Error("send response")
	}
}

func Decode(reader io.Reader, dst interface{}) error {
	if err := json.NewDecoder(reader).Decode(dst); err != nil {
		return fmt.Errorf("decode into %T: %w", dst, err)
	}
	return nil
}

func Encode(writer io.Writer, src interface{}) error {
	if err := json.NewEncoder(writer).Encode(src); err != nil {
		return fmt.Errorf("encode %T: %w", src, err)
	}
	return nil
}
