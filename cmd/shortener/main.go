package main

import (
	"github.com/size12/url-shortener/internal/app"
	"github.com/size12/url-shortener/internal/config"
	"log"
)

func main() {
	cfg := config.GetConfig()
	service := app.App{Cfg: cfg}
	log.Fatal(service.Run())
}
