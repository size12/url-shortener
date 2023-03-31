// Package config gets config.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
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

// GetTestConfig gets config for tests.
func GetTestConfig() Config {
	return Config{
		ServerAddress: ":8080",
		BaseURL:       "http://127.0.0.1:8081",
		StoragePath:   "file_storage.txt",
		BasePath:      "mockedDB",
		EnableHTTPS:   false,
	}
}

// GetBenchConfig gets config for benchmarks.
func GetBenchConfig() Config {
	return Config{
		DBMigrationPath: "file://../../migrations",
	}
}

// GetConfig gets new config from flags, env or file.
func GetConfig() Config {
	cfg := GetDefaultConfig()
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
	fmt.Println(cfg)
	cfg.ChangeByPriority(fileCfg)
	fmt.Println(cfg)
	cfg.ChangeByPriority(envCfg)
	fmt.Println(cfg)
	cfg.ChangeByPriority(flagCfg)
	fmt.Println(cfg)

	return cfg
}

var (
	cfg  = GetDefaultConfig()
	once sync.Once
)

// GetOldConfig gets new config from flags or env.
func GetOldConfig() Config {

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

// ChangeByPriority changes config by priority.
func (cfg *Config) ChangeByPriority(newCfg Config) {
	values := reflect.ValueOf(newCfg)
	oldValues := reflect.ValueOf(cfg).Elem()

	for j := 0; j < values.NumField(); j++ {
		if !values.Field(j).IsZero() {
			oldValues.Field(j).Set(values.Field(j))
		}
	}
}
