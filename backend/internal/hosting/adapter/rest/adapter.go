package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"vpainless/internal/hosting/core"
)

type Adapter struct {
	service HostingService
}

// HostingService encapsulates the core functionalities provided by
// instance core.
type HostingService interface {
	GetInstance(ctx context.Context, id core.InstanceID) (*core.Instance, error)
	DeleteInstance(ctx context.Context, id core.InstanceID) error
	CreateInstance(ctx context.Context) (*core.Instance, error)
	ListInstances(ctx context.Context) ([]*core.Instance, error)
}

// NewAdapter creates a new rest adapter to interact with hosting core
func NewAdapter(service HostingService) *Adapter {
	return &Adapter{service: service}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func toPointer[T comparable](v T) *T {
	var zero T
	if v == zero {
		return &zero
	}
	return &v
}
