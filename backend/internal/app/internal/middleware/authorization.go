package middleware

import (
	"context"
	"strings"
	"net/http"

	"github.com/OscarVillanueva/goapi/internal/app/tools"
)

// Define a complex string key to avoid collitions in the context
type contextKey string
const UserUUIDKey contextKey = "user_uuid"

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)  {
		authHeader := r.Header.Get("Authorization")

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		claims, err := tools.IsValidToken(parts[1])
		if err != nil {
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		data, ok := claims["data"].(map[string]any)
		if !ok {
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		uuid, ok := data["uuid"].(string)
		if !ok {
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		// Create a context based in the request context
		ctx := context.WithValue(r.Context(), UserUUIDKey, uuid)
		modified := r.WithContext(ctx)

		next.ServeHTTP(w, modified)
	})
}
