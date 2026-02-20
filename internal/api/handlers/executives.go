package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Executives endpoint accessed")
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET method in execs routes"))
	case http.MethodPost:
		fmt.Println("Query Parameters:", r.URL.Query())
		w.Write([]byte("Hello POST method in execs routes\n"))
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
