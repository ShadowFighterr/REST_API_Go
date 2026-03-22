package middlewares

import (
	"net/http"
	"strings"
)

func MiddlewaresExcludeRoutes(middleware func(http.Handler) http.Handler, excludedPaths []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		wrapped := middleware(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range excludedPaths {
				if strings.HasSuffix(path, "*") {
					prefix := strings.TrimSuffix(path, "*")
					if strings.HasPrefix(r.URL.Path, prefix) {
						next.ServeHTTP(w, r)
						return
					}
					continue
				}

				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			wrapped.ServeHTTP(w, r)
		})
	}
}
