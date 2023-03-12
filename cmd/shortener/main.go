package main

import (
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/url-shortener/internal/app"
	"github.com/size12/url-shortener/internal/config"
)

// Run example:  go run -ldflags "-X main.buildVersion=0.19 -X main.buildDate=12.03.23 -X main.buildCommit=iter19" cmd/shortener/main.go.

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	printBuildInfo()
	cfg := config.GetConfig()
	service := app.App{Cfg: cfg}
	log.Fatal(service.Run())
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)
}
