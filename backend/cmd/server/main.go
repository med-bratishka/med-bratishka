package main

import (
	"net/http"

	"medbratishka/internal/dependencies"
	"medbratishka/pkg/config"
	"medbratishka/pkg/logs"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()
	deps, err := dependencies.New(cfg)
	if err != nil {
		panic(err)
	}
	defer deps.Close()

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logs.CtxWithRequestID(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	deps.AuthHandler().FillHandlers(router)
	deps.BindingsHandler().FillHandlers(router)
	deps.ChatHandler().FillHandlers(router)

	deps.Logger().Infof("Server running on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		deps.Logger().Fatalf("failed to serve: %v", err)
	}
}
