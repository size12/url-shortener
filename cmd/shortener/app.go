package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/handlers"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"log"
	"net/http"
)

type App struct {
	cfg config.Config
}

func (app App) Run() error {
	r := chi.NewRouter()
	links, err := linkhelpers.NewStorage(app.cfg)
	if err != nil {
		log.Fatal(err)
	}
	server := http.Server{Addr: app.cfg.ServerAddress, Handler: r}
	r.Use(handlers.GzipHandle)
	r.Use(handlers.GzipRequest)
	r.MethodNotAllowed(handlers.URLErrorHandler)
	r.Get("/{id}", handlers.URLGetHandler(links))
	r.Post("/", handlers.URLPostHandler(links))
	r.Post("/api/shorten", handlers.URLPostHandler(links))
	return server.ListenAndServe()
}
