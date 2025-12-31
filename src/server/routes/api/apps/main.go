package appsApi

import "github.com/go-chi/chi/v5"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", getAppList)
	r.Get("/{appId}", getApp)
	r.Post("/", installApp)
	r.Delete("/{appId}", uninstallApp)

	return r
}
