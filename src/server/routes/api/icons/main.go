package iconsApi

import "github.com/go-chi/chi/v5"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/apps/{icon}", getAppIcon)

	return r
}
