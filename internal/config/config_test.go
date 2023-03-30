package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBenchConfig(t *testing.T) {
	cfg := GetBenchConfig()
	assert.Equal(t, Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://127.0.0.1:8081",
		StoragePath:     "file_storage.txt",
		BasePath:        "postgresql://",
		DBMigrationPath: "file://../../migrations",
		EnableHTTPS:     false,
	}, cfg)
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	assert.Equal(t, Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://127.0.0.1:8080",
		StoragePath:     "file_storage.txt",
		BasePath:        "postgresql://",
		EnableHTTPS:     false,
		DBMigrationPath: "file://migrations",
	}, cfg)
}

func TestGetConfig(t *testing.T) {
	os.Args = append(os.Args, "-a", ":9090", "-b", "https://127.0.0.1:9090", "-f", "file.txt", "-d", "", "-s", "-d", "postgresql://")
	cfg := GetConfig()
	assert.Equal(t, Config{
		ServerAddress:   ":9090",
		BaseURL:         "https://127.0.0.1:9090",
		StoragePath:     "file.txt",
		BasePath:        "postgresql://",
		EnableHTTPS:     true,
		DBMigrationPath: "file://migrations",
	}, cfg)
}

func TestChangeByPriority(t *testing.T) {
	cfg := GetTestConfig()
	newCfg := Config{BaseURL: "https://url-shortener.com"}
	cfg.changeByPriority(newCfg)

	assert.Equal(t, Config{
		ServerAddress:   cfg.ServerAddress,
		BaseURL:         "https://url-shortener.com",
		StoragePath:     cfg.StoragePath,
		BasePath:        cfg.BasePath,
		EnableHTTPS:     cfg.EnableHTTPS,
		DBMigrationPath: cfg.DBMigrationPath,
	}, cfg)
}
