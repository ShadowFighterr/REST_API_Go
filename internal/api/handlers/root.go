package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Root endpoint accessed")
	fmt.Fprint(w, "Hello, World!")
}
