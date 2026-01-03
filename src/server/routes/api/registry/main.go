package registryApi

import "github.com/go-chi/chi/v5"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", openRegistry)
	r.Delete("/", closeRegistry)

	return r
}
