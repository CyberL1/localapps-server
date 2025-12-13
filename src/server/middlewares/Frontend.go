//go:build !debug

package middlewares

import (
	"localapps-server/utils"
	"localapps-server/web"
	"net/http"
	"strings"
)

func FrontendProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == strings.Split(utils.ServerConfig.AccessUrl, "://")[1] && !strings.HasPrefix(r.URL.Path, "/api") {
			http.FileServerFS(web.BuildDirFS).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
