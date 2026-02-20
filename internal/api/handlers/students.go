package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func StudentsHandler(w http.ResponseWriter, r *http.Request) {
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
