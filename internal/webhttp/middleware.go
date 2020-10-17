package webhttp

import (
	"mime"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/go-chi/chi/middleware"

	"github.com/pavelmemory/jobtome/internal/logging"
)

// InjectLogger returns a middleware function that injects a logger into request's context.
// It also propagates logger with a request unique sequence number, so all the logs
// for a particular request could be grouped together.
func InjectLogger(logger logging.Logger) func(http.Handler) http.Handler {
	var reqSeq = new(int64)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.WithInt64("req_seq", atomic.AddInt64(reqSeq, 1))
			ctx := logging.ToContext(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LogRequest returns a middleware function that logs each incoming request.
// TODO: make logging level configurable so we could control log severity for each baseHandler
func LogRequest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logging.FromContext(r.Context())

			logger.WithString("url", r.URL.String()).
				WithString("method", r.Method).
				WithString("referer", r.Referer()).
				WithString("user_agent", r.UserAgent()).
				WithInt64("content_length", r.ContentLength).
				Debug("incoming request")

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				// if the status was not set explicitly (0 is default)
				// it is considered as OK by net/http
				status = http.StatusOK
			}

			logger.WithInt("status", status).
				WithInt("bytes_written", ww.BytesWritten()).
				Debug("outgoing response")
		})
	}
}

// AcceptsJSON verifies request has a 'content-type' header with 'application/json' mime type.
var AcceptsJSON = RequestContentType("application/json; charset=utf-8")

// RequestContentType returns a middleware function that verifies request has
// `content-type` header and its media type is equal to passed in value.
func RequestContentType(contentType string) func(http.Handler) http.Handler {
	wantMediaType, wantParams, err := mime.ParseMediaType(contentType)
	if err != nil {
		// this is fair enough as it will blow up at startup time
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType := r.Header.Get("content-type")
			mediaType, params, err := mime.ParseMediaType(contentType)
			if err != nil {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			if !strings.EqualFold(wantMediaType, mediaType) {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			for k, v := range params {
				if !strings.EqualFold(v, wantParams[k]) {
					w.WriteHeader(http.StatusUnsupportedMediaType)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ProducesJSON sets response header 'content-type' to with 'application/json' mime type.
var ProducesJSON = ResponseContentType("application/json; charset=utf-8")

// ResponseContentType returns a middleware function that sets passed in value
// as a `content-type` header to the HTTP response in case is not yet set.
func ResponseContentType(contentType string) func(http.Handler) http.Handler {
	_, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		// this is fair enough as it will blow up at startup time
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := &responseContentTypeWrapper{ResponseWriter: w, contentType: contentType}
			next.ServeHTTP(ww, r)
		})
	}
}

type responseContentTypeWrapper struct {
	contentType string
	http.ResponseWriter
}

func (rw *responseContentTypeWrapper) Write(d []byte) (int, error) {
	if rw.Header().Get("content-type") == "" && rw.contentType != "" {
		rw.ResponseWriter.Header().Set("content-type", rw.contentType)
	}

	return rw.ResponseWriter.Write(d)
}

func (rw *responseContentTypeWrapper) WriteHeader(statusCode int) {
	// keep plain response content-type for redirects (status: 300-308)
	if statusCode >= http.StatusMultipleChoices && statusCode <= http.StatusPermanentRedirect {
		rw.contentType = ""
	}

	rw.ResponseWriter.WriteHeader(statusCode)
}
