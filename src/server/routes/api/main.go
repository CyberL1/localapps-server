package api

import (
	"localapps-server/server/middlewares"
	adminApi "localapps-server/server/routes/api/admin"
	appsApi "localapps-server/server/routes/api/apps"
	iconsApi "localapps-server/server/routes/api/icons"

	"github.com/go-chi/chi/v5"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.ApiAuth)

		r.Mount("/admin", adminApi.NewHandler().RegisterRoutes())
		r.Mount("/apps", appsApi.NewHandler().RegisterRoutes())
	})

	r.Mount("/icons", iconsApi.NewHandler().RegisterRoutes())
	r.Get("/link", Link)

	return r
}
