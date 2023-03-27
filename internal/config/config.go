// Package config gets config.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/caarlos0/env/v6"
)

// Config Application config.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address,omitempty"`
	BaseURL         string `env:"BASE_URL" json:"base_url,omitempty"`
	StoragePath     string `env:"FILE_STORAGE_PATH" json:"storage_path,omitempty"`
	BasePath        string `env:"DATABASE_DSN" json:"base_path,omitempty"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https,omitempty"`
	DBMigrationPath string
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
		cfgFilePath := ""

		fileCfg := Config{}
		envCfg := Config{}
		flagCfg := Config{}

		// flag config.
		flag.StringVar(&flagCfg.ServerAddress, "a", "", "Server address")
		flag.StringVar(&flagCfg.BaseURL, "b", "", "Base URL")
		flag.StringVar(&flagCfg.StoragePath, "f", "", "Storage path")
		flag.StringVar(&flagCfg.BasePath, "d", "", "DataBase path")
		flag.BoolVar(&flagCfg.EnableHTTPS, "s", false, "Enable HTTPS")

		// file config.
		flag.StringVar(&cfgFilePath, "c", "", "Config file path")
		flag.StringVar(&cfgFilePath, "config", "", "Config file path")
		flag.Parse()

		if cfgFilePath != "" {
			file, err := os.ReadFile(cfgFilePath)
			if err != nil {
				log.Fatalln("Failed parse config file:", err)
			}

			err = json.Unmarshal(file, &fileCfg)
			if err != nil {
				log.Fatalln("Failed unmarshal config file:", err)
			}
		}

		// env config.

		err := env.Parse(&envCfg)
		if err != nil {
			log.Fatalln("Failed parse config:", err)
		}

		// change config by priority.
		changeByPriority(fileCfg)
		changeByPriority(envCfg)
		changeByPriority(flagCfg)
	})

	return cfg
}

// changeByPriority changes config by priority.
func changeByPriority(newCfg Config) {
	values := reflect.ValueOf(newCfg)
	oldValues := reflect.ValueOf(&cfg).Elem()

	for j := 0; j < values.NumField(); j++ {
		if !values.Field(j).IsZero() {
			oldValues.Field(j).Set(values.Field(j))
		}
	}
}
