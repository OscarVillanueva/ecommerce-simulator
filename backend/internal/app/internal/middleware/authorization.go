package middleware

import (
	"errors"
	"context"
	"strings"
	"net/http"

	"github.com/OscarVillanueva/goapi/internal/app/tools"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
)

// Define a complex string key to avoid collitions in the context
type contextKey string
const UserUUIDKey contextKey = "user_uuid"

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)  {
		tr := otel.Tracer("auth-middleware")

		trContext, span := tr.Start(r.Context(), "auth-handler")
		defer span.End()

		authHeader := r.Header.Get("Authorization")

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		claims, err := tools.IsValidToken(parts[1])
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Invalid Token")
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		data, ok := claims["data"].(map[string]any)
		if !ok {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Invalid Token, missing data claims")
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		uuid, ok := data["uuid"].(string)
		if !ok {
			span.RecordError(errors.New("Missin data.uuid"))
			span.SetStatus(codes.Error, "Invalid Token, missing data.uuid")
			tools.UnauthorizedErrorHandler(w, nil)
			return
		}

		// Create a context based in the request context
		ctx := context.WithValue(trContext, UserUUIDKey, uuid)
		modified := r.WithContext(ctx)

		span.SetStatus(codes.Ok, "Valid token")
		next.ServeHTTP(w, modified)
	})
}
