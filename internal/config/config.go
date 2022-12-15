package config

import (
	"github.com/caarlos0/env/v6"
	"os"
)

const (
	DefaultServerAddress string = ":8080"
	DefaultBaseURL       string = "http://127.0.0.1:8080"
	DefaultStoragePath   string = ""
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	StoragePath   string `env:"FILE_STORAGE_PATH" envDefault:""`
}

func findFlag(name string, defVal string) string {
	for i, val := range os.Args[1:] {
		if val == name {
			return os.Args[1:][i+1]
		}
	}
	return defVal
}

func GetConfig() Config {
	var cfg Config
	env.Parse(&cfg)
	aFlag := findFlag("-a", DefaultServerAddress)
	bFlag := findFlag("-b", DefaultBaseURL)
	fFlag := findFlag("-f", DefaultStoragePath)

	if cfg.ServerAddress != aFlag && aFlag != DefaultServerAddress {
		cfg.ServerAddress = aFlag
	}
	if cfg.BaseURL != bFlag && bFlag != DefaultBaseURL {
		cfg.BaseURL = bFlag
	}
	if cfg.StoragePath != fFlag && fFlag != DefaultStoragePath {
		cfg.StoragePath = fFlag
	}

	return cfg
}
