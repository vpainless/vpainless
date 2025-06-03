package middleware

import (
	"encoding/json"
	"net/http"
)

type MiddlewareFunc func(next http.Handler) http.Handler

type ctxtype string

const (
	basic     ctxtype = "basic"
	requestID ctxtype = "request_id"
)

func writeJSONError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
