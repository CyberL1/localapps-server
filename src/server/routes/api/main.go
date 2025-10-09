package api

import (
	"localapps-server/server/middlewares"
	adminApi "localapps-server/server/routes/api/admin"
	appsApi "localapps-server/server/routes/api/apps"
	iconsApi "localapps-server/server/routes/api/icons"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.Handle("/admin/", http.StripPrefix("/admin", middlewares.ApiAuth(adminApi.NewHandler().RegisterRoutes())))
	r.Handle("/apps/", http.StripPrefix("/apps", middlewares.ApiAuth(appsApi.NewHandler().RegisterRoutes())))
	r.Handle("/icons/", http.StripPrefix("/icons", iconsApi.NewHandler().RegisterRoutes()))

	r.HandleFunc("GET /link", Link)

	return r
}
