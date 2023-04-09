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
	"time"

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

	service := handlers.NewService(app.Cfg, s)

	server := &http.Server{
		Addr:      app.Cfg.ServerAddress,
		Handler:   r,
		TLSConfig: manager.TLSConfig(),
	}

	r.Use(handlers.CookieMiddleware)
	r.Use(handlers.GzipHandle)
	r.Use(handlers.GzipRequest)

	r.MethodNotAllowed(handlers.URLErrorHandler)
	r.Get("/ping", handlers.PingHandler(service))
	r.Get("/{id}", handlers.URLGetHandler(service))
	r.Get("/api/user/urls", handlers.URLHistoryHandler(service))
	r.Delete("/api/user/urls", handlers.DeleteHandler(service))
	r.Post("/", handlers.URLPostHandler(service))
	r.Post("/api/shorten/batch", handlers.URLBatchHandler(service))
	r.Post("/api/shorten", handlers.URLPostHandler(service))

	r.Group(func(r chi.Router) {
		r.Use(handlers.NewIPPermissionsChecker(app.Cfg))
		r.Get("/api/internal/stats", handlers.StatisticHandler(service))
	})

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigint
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
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
