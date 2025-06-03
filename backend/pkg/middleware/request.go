package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
)

// RequestIDMiddleware attaches a unique request id to each request context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), requestID, uuid.Must(uuid.NewV4()))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID returns the attached request id to the context.
func GetRequestID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(requestID)
	if val == nil {
		return uuid.Nil, fmt.Errorf("no request id on context")
	}

	requestID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid request id on context")
	}

	return requestID, nil
}
