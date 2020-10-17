package webhttp

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/pavelmemory/jobtome/internal/logging"
)

type Resolver interface {
	// Resolve returns a full URL accessioned with the hash.
	Resolve(ctx context.Context, hash string) (string, error)
}

func NewResolverHandler(resolver Resolver) ResolverHandler {
	return ResolverHandler{resolver: resolver}
}

// ResolverHandler redirects incoming requests by using request uri as a hash of the longer path.
type ResolverHandler struct {
	baseHandler
	resolver Resolver
}

func (rh ResolverHandler) Register(router chi.Router) {
	router = router.With(LogRequest())
	router.Method(http.MethodGet, "/{hash}", http.HandlerFunc(rh.Resolve))
}

func (rh ResolverHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := rh.logger(ctx, "Resolve")

	logger.Debug("start")
	defer logger.Debug("end")

	hash := rh.pathParam(r, "hash")
	url, err := rh.resolver.Resolve(r.Context(), hash)
	if err != nil {
		logger.WithError(err).WithString("hash", hash).Error("resolve hash")
		WriteError(w, logger, err)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (rh ResolverHandler) logger(ctx context.Context, method string) logging.Logger {
	return logging.FromContext(ctx).WithString("component", "ResolverHandler").WithString("method", method)
}
