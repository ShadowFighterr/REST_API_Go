package middlewares

import (
	"net/http"
	"strings"
)

type HPPOptions struct {
	CheckQuery                   bool
	CheckBody                    bool
	CheckBodyOnlyForContentTypes string
	Whitelist                    []string
}

func HPPMiddleware(options HPPOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.CheckBody && r.Method != http.MethodPost && isContentTypeAllowed(r, options.CheckBodyOnlyForContentTypes) {
				filterBodyParams(r, options.Whitelist)
			}
			if options.CheckQuery && r.URL.RawQuery != "" {
				filterQueryParams(r, options.Whitelist)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isContentTypeAllowed(r *http.Request, allowedTypes string) bool {
	return strings.Contains(allowedTypes, r.Header.Get("Content-Type"))
}

func filterBodyParams(r *http.Request, whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		return
	}
	for key, values := range r.Form {
		if len(values) > 1 {
			r.Form[key] = []string{values[0]}
		}
		if isWhitelisted(key, whitelist) {
			r.Form[key] = values
		}
	}
}

func filterQueryParams(r *http.Request, whitelist []string) {
	query := r.URL.Query()
	for key, values := range query {
		if len(values) > 1 && !isWhitelisted(key, whitelist) {
			query[key] = []string{values[0]}
		}
	}
	r.URL.RawQuery = query.Encode()
}

func isWhitelisted(key string, whitelist []string) bool {
	for _, w := range whitelist {
		if key == w {
			return true
		}
	}
	return false
}
