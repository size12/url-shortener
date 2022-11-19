package main

import (
	"github.com/size12/url-shortener/internal/handlers"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{Addr: "127.0.0.1:8080", Handler: mux}
	mux.HandleFunc("/", handlers.UrlHandler)
	log.Fatal(server.ListenAndServe())
}
