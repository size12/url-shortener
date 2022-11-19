package main

import (
	"github.com/size12/url-shortener/internal/handlers"
	"log"
	"net/http"
)

func main() {
	server := http.Server{Addr: "127.0.0.1:8080", Handler: http.HandlerFunc(handlers.URLHandler)}
	log.Fatal(server.ListenAndServe())
}
