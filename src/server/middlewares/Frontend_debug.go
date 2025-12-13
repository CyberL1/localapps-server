//go:build debug

package middlewares

import (
	"localapps-server/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func FrontendProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == strings.Split(utils.ServerConfig.AccessUrl, "://")[1] && !strings.HasPrefix(r.URL.Path, "/api") {
			frontendUrl, _ := url.Parse("http://localhost:5173")
			httputil.NewSingleHostReverseProxy(frontendUrl).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
