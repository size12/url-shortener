package config

import "github.com/caarlos0/env/v6"

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
}

func GetConfig() Config {
	var cfg Config
	env.Parse(&cfg)
	return cfg
}
