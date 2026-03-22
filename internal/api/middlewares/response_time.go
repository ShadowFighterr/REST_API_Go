package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func ResponseTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Response Time Middleware triggered")
		start := time.Now()

		//Create a custom response writer to capture status code if needed
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)
		duration := time.Since(start)
		wrappedWriter.Header().Set("X-Response-Time", duration.String())

		fmt.Printf("Method: %s, Path: %s, Duration: %v, Status: %d\n", r.Method, r.URL.Path, duration, wrappedWriter.statusCode)
		fmt.Println("Response Time Middleware completed")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
