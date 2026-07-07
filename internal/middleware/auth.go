package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ezhigval/blog-cms-api/internal/auth"
	"github.com/ezhigval/go-toolkit/httputil"
)

func Authenticate(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httputil.WriteError(w, httputil.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil))
				return
			}
			claims, err := tokens.Parse(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				httputil.WriteError(w, httputil.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", err))
				return
			}
			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (int64, bool) {
	c, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return 0, false
	}
	return c.UserID, true
}
