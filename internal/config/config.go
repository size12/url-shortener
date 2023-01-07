package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	StoragePath   string `env:"FILE_STORAGE_PATH"`
	BasePath      string `env:"DATABASE_DSN"`
}

func GetDefaultConfig() Config {
	return Config{
		ServerAddress: ":8080",
		BaseURL:       "http://127.0.0.1:8080",
		StoragePath:   "",
		BasePath:      "",
	}
}

func GetConfig() Config {
	cfg := GetDefaultConfig()
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.StoragePath, "f", cfg.StoragePath, "Storage path")
	flag.StringVar(&cfg.BasePath, "d", cfg.BasePath, "DataBase path")
	flag.Parse()

	env.Parse(&cfg)

	return cfg
}
