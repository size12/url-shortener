package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/handlers"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	cfg := config.GetConfig()
	links, err := linkhelpers.NewStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}
	server := http.Server{Addr: cfg.ServerAddress, Handler: r}
	r.Use(handlers.MiddlewareGzip)
	r.MethodNotAllowed(handlers.URLErrorHandler)
	r.Get("/{id}", handlers.URLGetHandler(cfg, links))
	r.Post("/", handlers.URLPostHandler(cfg, links))
	r.Post("/api/shorten", handlers.URLPostHandler(cfg, links))
	log.Fatal(server.ListenAndServe())
}
