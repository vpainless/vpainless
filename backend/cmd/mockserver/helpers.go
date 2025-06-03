package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
)

func basicAuth(r *http.Request) (user, password string, err error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		err = errors.New("Invalid authorization header")
		return
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		err = errors.New(`Only "Basic" authorization is allowed`)
		return
	}

	b, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		err = errors.New(fmt.Sprintf("invalid token %s: %v", parts[1], err))
		return
	}

	creds := strings.Split(string(b), ":")
	if len(creds) != 2 {
		err = errors.New("Invalid authorization header")
		return
	}

	return creds[0], creds[1], nil
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
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
