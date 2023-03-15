package storage

import (
	"testing"

	"github.com/size12/url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDBStorage(t *testing.T) {
	cfg := config.Config{BaseURL: "postgres://127.0.0.1:5432", DBMigrationPath: "file://../../migrations"}
	_, err := NewDBStorage(cfg)
	if err != nil {
		t.Log("Please run DB. Can't run tests.")
		return
	}

	assert.NoError(t, err, "Create new DB storage.")
}

func TestDBStorage(t *testing.T) {
	cfg := config.Config{BaseURL: "http://127.0.0.1", BasePath: "postgres://127.0.0.1:5432", DBMigrationPath: "file://../../migrations"}
	s, err := NewDBStorage(cfg)
	if err != nil {
		t.Log("Please run DB. Can't run tests.")
		return
	}
	// Get config.
	assert.Equal(t, cfg, s.GetConfig())

	// Ping DB.
	err = s.Ping()
	assert.NoError(t, err)

	// Creating short urls.
	res, err := s.CreateShort("user12", "https://yandex.ru", "https://google.com")
	if err != nil && err != Err409 {
		t.Error("Failed create short urls: ", err)
		return
	}

	// Getting long url.
	long, err := s.GetLong(res[0])
	if err != nil && err != Err404 {
		t.Error("Failed create short urls: ", err)
		return
	}

	assert.Equal(t, "https://yandex.ru", long)
}
