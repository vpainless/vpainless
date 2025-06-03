package authz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"vpainless/pkg/middleware"
)

var ErrPrincipalNotFound = errors.New("no principal was found on context")

type Authenticator interface {
	Authenticate(ctx context.Context, cred middleware.Credentials) (Principal, error)
}

// AuthenticationMiddleware authenticates users with their provided creds.
// If their creds are valid, a principal is set on the context, otherwise
// the request is rejected.
func AuthenticationMiddleware(auth Authenticator, exclusions []middleware.Exclusion) middleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			excluded := false
			for _, e := range exclusions {
				// TODO: use regex instead of prefix
				if strings.HasPrefix(r.URL.Path, e.PathPrefix) && r.Method == e.Method {
					excluded = true
				}
			}

			ctx := r.Context()
			creds, err := middleware.GetCreds(ctx)
			if err != nil {
				if excluded {
					next.ServeHTTP(w, r)
					return
				}
				writeJSONError(w, err)
				return
			}

			principal, err := auth.Authenticate(ctx, creds)
			if err != nil {
				writeJSONError(w, fmt.Errorf("invalid username or password: %w", err))
				return
			}

			ctx = context.WithValue(r.Context(), authz, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetPrincipal retrieves the attached principal to the request context
func GetPrincipal(ctx context.Context) (Principal, error) {
	val := ctx.Value(authz)
	if val == nil {
		return Principal{}, ErrPrincipalNotFound
	}

	principal, ok := val.(Principal)
	if !ok {
		return Principal{}, ErrPrincipalNotFound
	}

	return principal, nil
}

func writeJSONError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
