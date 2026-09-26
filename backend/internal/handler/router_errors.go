package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"medbratishka/pkg/logger"

	"github.com/gorilla/mux"
)

func ApplyRouterErrorHandlers(router *mux.Router, log logger.Logger) {
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		makeErrorResponse(w, r, log, http.StatusNotFound, "ROUTE_NOT_FOUND", "route not found", nil)
	})
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		makeErrorResponse(w, r, log, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	})
}

func RecoveryMiddleware(log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if log != nil {
						log.Errorf("panic recovered: %v stack: %s", rec, debug.Stack())
					}
					makeErrorResponse(w, r, log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", fmt.Errorf("panic: %v", rec))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
