package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

var cfgFlags Config

const (
	DefaultServerAddress string = ":8080"
	DefaultBaseURL       string = "http://127.0.0.1:8080"
	DefaultStoragePath   string = ""
)

func init() {
	flag.StringVar(&cfgFlags.ServerAddress, "a", DefaultServerAddress, "Server address")
	flag.StringVar(&cfgFlags.BaseURL, "b", DefaultBaseURL, "Base URL")
	flag.StringVar(&cfgFlags.StoragePath, "f", DefaultStoragePath, "Storage Path")
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

	if cfg.ServerAddress != cfgFlags.ServerAddress && cfgFlags.ServerAddress != DefaultServerAddress {
		cfg.ServerAddress = cfgFlags.ServerAddress
	}
	if cfg.BaseURL != cfgFlags.BaseURL && cfgFlags.BaseURL != DefaultBaseURL {
		cfg.BaseURL = cfgFlags.BaseURL
	}
	if cfg.StoragePath != cfgFlags.StoragePath && cfgFlags.StoragePath != DefaultStoragePath {
		cfg.StoragePath = cfgFlags.StoragePath
	}

	return cfg
}
