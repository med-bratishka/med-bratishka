package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"medbratishka/internal/domain"
	"medbratishka/internal/service"
	"medbratishka/pkg/logger"
)

const (
	contextKeyUser = "user"
)

func AuthMiddleware(authService service.AuthService, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				makeErrorResponse(w, r, log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", fmt.Errorf("missing or malformed jwt"))
				return
			}

			userCtx, err := authService.ValidateToken(r.Context(), domain.PurposeAccess, tokenString)
			if err != nil {
				makeErrorResponse(w, r, log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", err)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUser, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RefreshMiddleware(authService service.AuthService, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				makeErrorResponse(w, r, log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", fmt.Errorf("missing or malformed jwt"))
				return
			}

			userCtx, err := authService.ValidateToken(r.Context(), domain.PurposeRefresh, tokenString)
			if err != nil {
				makeErrorResponse(w, r, log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", err)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUser, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRolesMiddleware(log logger.Logger, roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserFromContext(r)
			if userCtx == nil {
				makeErrorResponse(w, r, log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
				return
			}
			if _, ok := allowed[userCtx.Role]; !ok {
				makeErrorResponse(w, r, log, http.StatusForbidden, "FORBIDDEN", "forbidden", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	header := r.Header.Get(headerAuthorization)
	if strings.HasPrefix(header, bearerPrefix) {
		return header[len(bearerPrefix):]
	}
	return ""
}

// GetUserFromContext извлекает контекст пользователя из request
func GetUserFromContext(r *http.Request) *domain.UserTokenContext {
	userCtx, ok := r.Context().Value(contextKeyUser).(*domain.UserTokenContext)
	if !ok {
		return nil
	}
	return userCtx
}
