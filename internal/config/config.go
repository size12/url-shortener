package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	StoragePath   string `env:"FILE_STORAGE_PATH" envDefault:""`
}

func GetConfig() Config {
	var cfg Config
	env.Parse(&cfg)
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.StoragePath, "f", cfg.StoragePath, "Storage Path")
	flag.Parse()
	return cfg
}
