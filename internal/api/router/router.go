package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {

	// mux := http.NewServeMux()
	// mux.HandleFunc("GET /", handlers.RootHandler)

	tRouter := teachersRouter()
	sRouter := studentsRouter()

	sRouter.Handle("/", execsRouter())
	tRouter.Handle("/", sRouter)
	return tRouter

	// return mux
}
