package webhttp

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type baseHandler struct{}

func (uh baseHandler) pathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func (uh baseHandler) pathParamInt64(r *http.Request, opts ParamInt64Opts) (int64, error) {
	val := uh.pathParam(r, opts.P.Name)
	return opts.parse(val)
}

func (uh baseHandler) queryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

type ParamOpts struct {
	Optional bool
	Name     string
}

type ParamInt64Opts struct {
	P       ParamOpts
	Default int64
}

func (opts ParamInt64Opts) parse(val string) (int64, error) {
	if val == "" {
		if opts.P.Optional {
			return opts.Default, nil
		}
		return 0, ErrMissingRequired
	}

	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrBadFormat, err)
	}

	return parsed, nil
}

func (uh baseHandler) queryParamInt64(r *http.Request, opts ParamInt64Opts) (int64, error) {
	val := uh.queryParam(r, opts.P.Name)
	return opts.parse(val)
}
