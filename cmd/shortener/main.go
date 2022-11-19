package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/handlers"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	server := http.Server{Addr: "127.0.0.1:8080", Handler: r}
	r.MethodNotAllowed(handlers.ErrorHandler)
	r.Get("/{id}", handlers.URLGetHandler)
	r.Post("/", handlers.URLPostHandler)
	log.Fatal(server.ListenAndServe())
}
