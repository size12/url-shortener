package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/handlers"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"log"
	"net/http"
)

type App struct {
	Cfg config.Config
}

func (app App) Run() error {
	r := chi.NewRouter()
	links, err := linkhelpers.NewStorage(app.Cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer links.DB.Close()
	server := http.Server{Addr: app.Cfg.ServerAddress, Handler: r}
	r.Use(handlers.CookieMiddleware)
	r.Use(handlers.GzipHandle)
	r.Use(handlers.GzipRequest)
	r.MethodNotAllowed(handlers.URLErrorHandler)
	r.Get("/ping", handlers.PingHandler(links))
	r.Get("/{id}", handlers.URLGetHandler(links))
	r.Get("/api/user/urls", handlers.URLHistoryHandler(links))
	r.Post("/", handlers.URLPostHandler(links))
	r.Post("/api/shorten/batch", handlers.URLBatchHandler(links))
	r.Post("/api/shorten", handlers.URLPostHandler(links))
	return server.ListenAndServe()
}
