// Package app - package that creates service and runs it.
package app

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/handlers"
	"github.com/size12/url-shortener/internal/storage"
	"golang.org/x/crypto/acme/autocert"
)

// App is struct of service.
type App struct {
	Cfg config.Config
}

// Run starts service.
func (app App) Run() {
	r := chi.NewRouter()
	s, err := storage.NewStorage(app.Cfg)

	if err != nil {
		log.Fatal(err)
	}

	baseURL, err := url.Parse(app.Cfg.BaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	manager := &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(baseURL.Host),
	}

	server := &http.Server{
		Addr:      app.Cfg.ServerAddress,
		Handler:   r,
		TLSConfig: manager.TLSConfig(),
	}

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

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigint
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println("Failed shutdown server:", err)
		}
		close(idleConnsClosed)
	}()

	if app.Cfg.EnableHTTPS {
		if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServeTLS error: %v", err)
		}
	}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe error: %v", err)
	}

	<-idleConnsClosed
	log.Println("Shutdown server gracefully.")
}
