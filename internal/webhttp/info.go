package webhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/pavelmemory/jobtome/internal"
)

// InfoHandler handles requests about service status.
type InfoHandler struct{}

// Register creates a binding between method handlers and endpoints.
func (ih InfoHandler) Register(router chi.Router) {
	router.Method(http.MethodGet, "/-/liveness", http.HandlerFunc(ih.Readiness))
	router.Method(http.MethodGet, "/-/readiness", http.HandlerFunc(ih.Liveness))
	router.With(ProducesJSON).Method(http.MethodGet, "/-/version", http.HandlerFunc(ih.Version))
}

// Liveness returns HTTP status `200` for each request.
// It is used to determine if the service is healthy.
func (InfoHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Readiness returns HTTP status `200` for each request.
// It is used to determine if the service is ready to start receiving requests.
func (InfoHandler) Readiness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Version return information about binary: version, commit sha, timestamp of compilation
func (InfoHandler) Version(w http.ResponseWriter, _ *http.Request) {
	_ = internal.WriteVersion(json.NewEncoder(w))
}
