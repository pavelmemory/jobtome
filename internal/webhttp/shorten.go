package webhttp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/pavelmemory/jobtome/internal/logging"
	"github.com/pavelmemory/jobtome/internal/shorten"
)

//go:generate mockgen -source=shorten.go -destination mock.go -package webhttp ShortenService

// ShortenService provides set of operations available to operate on the user entity.
type ShortenService interface {
	// Create creates a new shorten and returns its unique identifier.
	Create(ctx context.Context, entity shorten.Entity) (int64, error)
	// Get returns a single shorten by its unique identifier.
	Get(ctx context.Context, id int64) (shorten.Entity, error)
	// List returns subset of all shortens.
	List(ctx context.Context, pager shorten.Pager) ([]shorten.Entity, error)
	// Delete removes shorten by its unique identifier.
	Delete(ctx context.Context, id int64) error
	// Resolve returns a full URL accessioned with the hash.
	Resolve(ctx context.Context, hash string) (string, error)
}

// NewShortenHandler returns HTTP baseHandler initialized with provided service abstraction.
func NewShortenHandler(shortenService ShortenService) ShortenHandler {
	return ShortenHandler{shortenService: shortenService}
}

// ShortenHandler handles request for the user entity(-ies).
type ShortenHandler struct {
	baseHandler
	shortenService ShortenService
	mapper         Mapper
}

// Register creates a binding between method handlers and endpoints.
func (uh ShortenHandler) Register(router chi.Router) {
	router = router.With(LogRequest())
	router.With(ProducesJSON, AcceptsJSON).Method(http.MethodPost, uh.urlPrefix(), http.HandlerFunc(uh.Create))
	router.With(ProducesJSON).Method(http.MethodGet, uh.urlPrefix(), http.HandlerFunc(uh.List))
	router.With(ProducesJSON).Method(http.MethodGet, uh.urlPrefix()+"/{id}", http.HandlerFunc(uh.Get))
	router.Method(http.MethodDelete, uh.urlPrefix()+"/{id}", http.HandlerFunc(uh.Delete))
}

func (uh ShortenHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "Create")

	logger.Debug("start")
	defer logger.Debug("end")

	var req CreateShortenReq
	if err := Decode(r.Body, &req); err != nil {
		logger.WithError(err).Error("decode payload")
		ErrorResponse{Cause: err, StatusCode: http.StatusBadRequest}.Write(logger, w)
		return
	}

	id, err := uh.shortenService.Create(ctx, uh.mapper.createShortenReq2Entity(req))
	if err != nil {
		logger.WithError(err).Error("creation of the shorten")
		WriteError(w, logger, err)
		return
	}

	w.Header().Set("location", uh.urlPrefix()+"/"+strconv.FormatInt(id, 10))
	w.WriteHeader(http.StatusCreated)
}

func (uh ShortenHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "Get")

	logger.Debug("start")
	defer logger.Debug("end")

	id, err := uh.pathParamInt64(r, ParamInt64Opts{P: ParamOpts{Name: "id"}})
	if err != nil {
		cause := fmt.Errorf(`parameter "id": %w`, err)
		logger.WithError(cause).Error("extract path parameter")
		ErrorResponse{Cause: cause, StatusCode: http.StatusBadRequest}.Write(logger, w)
		return
	}

	entity, err := uh.shortenService.Get(ctx, id)
	if err != nil {
		logger.WithError(err).WithInt64("id", id).Error(`get shorten by "id"`)
		WriteError(w, logger, err)
		return
	}

	if err := Encode(w, uh.mapper.entity2GetShortenResp(entity)); err != nil {
		logger.WithError(err).Error("encode entity")
		ErrorResponse{Cause: err, StatusCode: http.StatusInternalServerError}.Write(logger, w)
		return
	}
}

const (
	defaultListLimit = int64(50)
)

func (uh ShortenHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "List")

	logger.Debug("start")
	defer logger.Debug("end")

	limit, err := uh.queryParamInt64(r, ParamInt64Opts{P: ParamOpts{Name: "limit", Optional: true}, Default: defaultListLimit})
	if err != nil {
		cause := fmt.Errorf(`parameter "limit": %w`, err)
		logger.WithError(cause).Error("extract query parameter")
		ErrorResponse{Cause: cause, StatusCode: http.StatusBadRequest}.Write(logger, w)
		return
	}

	offset, err := uh.queryParamInt64(r, ParamInt64Opts{P: ParamOpts{Name: "offset", Optional: true}})
	if err != nil {
		cause := fmt.Errorf(`parameter "offset": %w`, err)
		logger.WithError(cause).Error("extract query parameter")
		ErrorResponse{Cause: cause, StatusCode: http.StatusBadRequest}.Write(logger, w)
		return
	}

	entities, err := uh.shortenService.List(ctx, shorten.Pager{Limit: limit, Offset: offset})
	if err != nil {
		logger.WithError(err).Error("extract shortens")
		WriteError(w, logger, err)
		return
	}

	if err := Encode(w, uh.mapper.entities2ListShortenResp(entities)); err != nil {
		logger.WithError(err).Error("encode shortens")
		ErrorResponse{Cause: err, StatusCode: http.StatusInternalServerError}.Write(logger, w)
		return
	}
}

func (uh ShortenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := uh.logger(ctx, "Delete")

	logger.Debug("start")
	defer logger.Debug("end")

	id, err := uh.pathParamInt64(r, ParamInt64Opts{P: ParamOpts{Name: "id"}})
	if err != nil {
		cause := fmt.Errorf(`parameter "id": %w`, err)
		logger.WithError(cause).Error("extract path parameter")
		ErrorResponse{Cause: cause, StatusCode: http.StatusBadRequest}.Write(logger, w)
		return
	}

	if err := uh.shortenService.Delete(ctx, id); err != nil {
		logger.WithError(err).WithInt64("id", id).Error("delete shorten")
		WriteError(w, logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (uh ShortenHandler) urlPrefix() string {
	return "/api/shorten"
}

func (uh ShortenHandler) logger(ctx context.Context, method string) logging.Logger {
	return logging.FromContext(ctx).WithString("component", "ShortenHandler").WithString("method", method)
}
