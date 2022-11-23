package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/handlers"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"log"
	"net/http"
	"sync"
)

func main() {
	r := chi.NewRouter()
	links := linkhelpers.URLLinks{Locations: make(map[string]string), Mutex: &sync.Mutex{}}
	server := http.Server{Addr: "127.0.0.1:8080", Handler: r}
	r.MethodNotAllowed(handlers.URLErrorHandler)
	r.Get("/{id}", handlers.URLGetHandler(links))
	r.Post("/", handlers.URLPostHandler(links))
	log.Fatal(server.ListenAndServe())
}
