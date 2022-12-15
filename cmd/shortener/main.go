package main

import (
	"github.com/size12/url-shortener/internal/config"
	"log"
)

func main() {
	cfg := config.GetConfig()
	app := App{cfg}
	log.Fatal(app.Run())
}
