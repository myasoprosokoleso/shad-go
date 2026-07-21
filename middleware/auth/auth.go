//go:build !solution

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

const userKey ctxKey = iota

type User struct {
	Name  string
	Email string
}

type ctxKey int

func ContextUser(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userKey).(*User)
	return user, ok
}

var ErrInvalidToken = errors.New("invalid token")

type TokenChecker interface {
	CheckToken(ctx context.Context, token string) (*User, error)
}

func CheckAuth(checker TokenChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header (want Authorization: Bearer <token>)", http.StatusUnauthorized)
				return
			}
			token, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found || token == "" {
				http.Error(w, "invalid authorization header (want Authorization: Bearer <token>)", http.StatusUnauthorized)
				return
			}

			user, err := checker.CheckToken(r.Context(), token)
			if err != nil || user == nil {
				if errors.Is(err, ErrInvalidToken) {
					http.Error(w, "invalid token", http.StatusUnauthorized)
				} else {
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
