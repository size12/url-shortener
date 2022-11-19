package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{Addr: "127.0.0.1:8080", Handler: mux}
	err := server.ListenAndServe()
	log.Fatal(err)
}
