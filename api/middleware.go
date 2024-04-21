package api

import (
	"context"
	"net/http"
)

type requestContextKey struct {
	Key string `json:"key"`
}

func (app *application) WithScope(scope []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, t, err := app.authenticateToken(r, scope)
			if err != nil {
				app.invalidCredentials(w, err)
				return
			}

			// Add user  and token to context
			ctx := context.WithValue(r.Context(), requestContextKey{Key: "user"}, u)
			ctx = context.WithValue(ctx, requestContextKey{Key: "token"}, t)
			r2 := r.WithContext(ctx)
			next.ServeHTTP(w, r2)
		})
	}
}
