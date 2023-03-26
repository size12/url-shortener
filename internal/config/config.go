// Package config gets config.
package config

import (
	"flag"
	"log"
	"sync"

	"github.com/caarlos0/env/v6"
)

// Config Application config.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	StoragePath     string `env:"FILE_STORAGE_PATH"`
	BasePath        string `env:"DATABASE_DSN"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
	DBMigrationPath string
	*sync.Once
}

// GetDefaultConfig gets default config.
func GetDefaultConfig() Config {
	return Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://127.0.0.1:8080",
		StoragePath:     "",
		BasePath:        "",
		DBMigrationPath: "file://migrations",
		EnableHTTPS:     false,
	}
}

// Singleton config creation variables.
var (
	cfg  = GetDefaultConfig()
	once sync.Once
)

// GetConfig gets new config from flags or env.
func GetConfig() Config {

	once.Do(func() {
		flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Server address")
		flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
		flag.StringVar(&cfg.StoragePath, "f", cfg.StoragePath, "Storage path")
		flag.StringVar(&cfg.BasePath, "d", cfg.BasePath, "DataBase path")
		flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS")
		flag.Parse()

		err := env.Parse(&cfg)
		if err != nil {
			log.Fatalln("Failed parse config: ", err)
		}
	})

	return cfg
}
