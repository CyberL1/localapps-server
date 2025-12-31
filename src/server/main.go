package server

import (
	"fmt"
	"localapps-server/constants"
	"localapps-server/server/middlewares"
	"localapps-server/server/routes/api"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Start() {
	router := chi.NewRouter()

	if constants.IsDebugBuild {
		router.Use(middleware.Logger)
	}

	router.Use(middleware.RedirectSlashes)

	router.Mount("/api", api.NewHandler().RegisterRoutes())

	if err := http.ListenAndServe(":8080", middlewares.FrontendProxy(middlewares.AppProxy(router))); err != nil {
		fmt.Printf("Failed to bind to port 8080: %s\n", err)
		os.Exit(1)
	}
}
