package app

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/handlers"
	"github.com/size12/url-shortener/internal/storage"
)

type App struct {
	Cfg config.Config
}

func (app App) Run() error {
	fmem, err := os.Create(`profiles/base.pprof`)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := pprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	s, err := storage.NewStorage(app.Cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := http.Server{Addr: app.Cfg.ServerAddress, Handler: r}

	r.Use(handlers.CookieMiddleware)
	r.Use(handlers.GzipHandle)
	r.Use(handlers.GzipRequest)

	r.MethodNotAllowed(handlers.URLErrorHandler)
	r.Get("/ping", handlers.PingHandler(s))
	r.Get("/{id}", handlers.URLGetHandler(s))
	r.Get("/api/user/urls", handlers.URLHistoryHandler(s))
	r.Delete("/api/user/urls", handlers.DeleteHandler(s))
	r.Post("/", handlers.URLPostHandler(s))
	r.Post("/api/shorten/batch", handlers.URLBatchHandler(s))
	r.Post("/api/shorten", handlers.URLPostHandler(s))

	return server.ListenAndServe()
}
