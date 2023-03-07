package main

import (
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/url-shortener/internal/app"
	"github.com/size12/url-shortener/internal/config"
)

func main() {
	cfg := config.GetConfig()
	service := app.App{Cfg: cfg}
	log.Fatal(service.Run())
}
