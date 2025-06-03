package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Credentials struct {
	Username string
	Password string
}

var ErrCredsNotFoundOnContext = errors.New("credntials not found on the context")

// Exclusion is used to prevent middlewares from being applied to certain paths.
type Exclusion struct {
	PathPrefix string
	Method     string
}

// BasicAuthMiddleware is a middleware to ensure that authorization header
// is present and is of type basic. It extracts the username and password
// and sets them on the request context.
func BasicAuthMiddleware(exclusions []Exclusion) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			excluded := false
			for _, e := range exclusions {
				// TODO: use regex instead of prefix
				if strings.HasPrefix(r.URL.Path, e.PathPrefix) && r.Method == e.Method {
					excluded = true
				}
			}
			auth := r.Header.Get("Authorization")
			if auth == "" {
				if excluded {
					next.ServeHTTP(w, r)
					return
				}
				writeJSONError(w, "Invalid authorization header")
				return
			}

			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				writeJSONError(w, `Only "Basic" authorization is allowed`)
				return
			}

			b, err := base64.URLEncoding.DecodeString(parts[1])
			if err != nil {
				writeJSONError(w, fmt.Sprintf("invalid token %s: %v", parts[1], err))
				return
			}

			creds := strings.Split(string(b), ":")
			if len(creds) != 2 {
				writeJSONError(w, "Invalid authorization header")
				return
			}

			ctx := context.WithValue(r.Context(), basic, Credentials{
				Username: creds[0],
				Password: creds[1],
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetCreds retrieves the attached creds to the request context
func GetCreds(ctx context.Context) (Credentials, error) {
	val := ctx.Value(basic)
	if val == nil {
		return Credentials{}, ErrCredsNotFoundOnContext
	}

	creds, ok := val.(Credentials)
	if !ok {
		return Credentials{}, ErrCredsNotFoundOnContext
	}

	return creds, nil
}
