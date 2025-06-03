package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid/v5"
)

type AccessService interface {
	userService
	groupService
}

type Adapter struct {
	service AccessService
}

func NewAdapter(service AccessService) *Adapter {
	return &Adapter{service}
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

func fromPointer[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func toPointer[T comparable](v T) *T {
	var zero T
	if v == zero {
		return &zero
	}
	return &v
}

func toUUIDPointer(id uuid.UUID) *uuid.UUID {
	if id.IsNil() {
		return nil
	}

	return &id
}
