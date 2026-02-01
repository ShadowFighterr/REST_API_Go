package middlewares

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Powered-By", "PHP 7.4.3")        // Example of obscuring server technology
		w.Header().Set("Server", "Apache/2.4.41 (Ubuntu)") // Example of obscuring server technology
		// w.Header().Set("Permissions-Policy", "geolocation=(), microphone=()") // Example of controlling feature policies
		next.ServeHTTP(w, r)
	})
}
