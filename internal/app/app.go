package app

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/handlers"
	"github.com/size12/url-shortener/internal/storage"
)

type App struct {
	Cfg config.Config
}

func (app App) Run() error {
	r := chi.NewRouter()
	storage, err := storage.NewStorage(app.Cfg)
	if err != nil {
		log.Fatal(err)
	}
	//defer links.DB.Close()
	server := http.Server{Addr: app.Cfg.ServerAddress, Handler: r}
	//go links.DBDeleteURLs()

	r.Use(handlers.CookieMiddleware)
	r.Use(handlers.GzipHandle)
	r.Use(handlers.GzipRequest)
	r.MethodNotAllowed(handlers.URLErrorHandler)
	//r.Get("/ping", handlers.PingHandler(links))
	r.Get("/{id}", handlers.URLGetHandler(storage))
	r.Get("/api/user/urls", handlers.URLHistoryHandler(storage))
	r.Delete("/api/user/urls", handlers.DeleteHandler(storage))
	r.Post("/", handlers.URLPostHandler(storage))
	r.Post("/api/shorten/batch", handlers.URLBatchHandler(storage))
	r.Post("/api/shorten", handlers.URLPostHandler(storage))
	return server.ListenAndServe()
}
