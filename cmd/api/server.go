package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	mw "restapi/internal/api/middlewares"
	"time"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Root endpoint accessed")
	fmt.Fprint(w, "Hello, World!")
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Students endpoint accessed")
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET method in students routes"))
	case http.MethodPost:
		w.Write([]byte("Hello POST method in students routes"))
	case http.MethodPut:
		w.Write([]byte("Hello PUT method in students routes"))
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH method in students routes"))
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE method in students routes"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	fmt.Fprint(w, "Students Endpoint")
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Teachers endpoint accessed")
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET method in teachers routes"))
	case http.MethodPost:
		w.Write([]byte("Hello POST method in teachers routes"))
	case http.MethodPut:
		w.Write([]byte("Hello PUT method in teachers routes"))
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH method in teachers routes"))
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE method in teachers routes"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	fmt.Fprint(w, "Teachers Endpoint")
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Executives endpoint accessed")
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET method in execs routes"))
	case http.MethodPost:
		fmt.Println("Query Parameters:", r.URL.Query())

		w.Write([]byte("Hello POST method in execs routes"))
	case http.MethodPut:
		w.Write([]byte("Hello PUT method in execs routes"))
	case http.MethodPatch:
		w.Write([]byte("Hello PATCH method in execs routes"))
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE method in execs routes"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	fmt.Fprint(w, "Executives Endpoint")
}

func applyMiddlewares(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func main() {
	port := ":3000"

	cert := "cert.pem"
	key := "key.pem"

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)

	mux.HandleFunc("/students/", studentsHandler)

	mux.HandleFunc("/teachers/", teachersHandler)

	mux.HandleFunc("/execs/", execsHandler)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := mw.NewRateLimiter(100, 1*time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery:                   true,
		CheckBody:                    true,
		CheckBodyOnlyForContentTypes: "application/x-www-form-urlencoded",
		Whitelist:                    []string{"SortBy", "Asc"},
	}
	secureMux := applyMiddlewares(mux,
		mw.SecurityHeaders,
		mw.Cors,
		rl.Limit,
		mw.HPPMiddleware(hppOptions),
		mw.ResponseTimeMiddleware,
		mw.CompressionMiddleware,
	)

	server := &http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("Server is running on port %s\n", port)
	if err := server.ListenAndServeTLS(cert, key); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
