package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

var cfgFlags Config

func init() {
	flag.StringVar(&cfgFlags.ServerAddress, "a", ":8080", "Server address")
	flag.StringVar(&cfgFlags.BaseURL, "b", "http://127.0.0.1:8080", "Base URL")
	flag.StringVar(&cfgFlags.StoragePath, "f", "", "Storage Path")
}

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	StoragePath   string `env:"FILE_STORAGE_PATH" envDefault:""`
}

func GetConfig() Config {
	var cfg Config
	env.Parse(&cfg)

	flag.Parse()

	if cfg.ServerAddress != cfgFlags.ServerAddress {
		cfg.ServerAddress = cfgFlags.ServerAddress
	}
	if cfg.BaseURL != cfgFlags.BaseURL {
		cfg.BaseURL = cfgFlags.BaseURL
	}
	if cfg.StoragePath != cfgFlags.StoragePath {
		cfg.StoragePath = cfgFlags.StoragePath
	}

	return cfg
}
