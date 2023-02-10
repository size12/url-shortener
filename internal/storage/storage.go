package storage

import (
	"errors"

	"github.com/size12/url-shortener/internal/config"
)

var Err409 = errors.New("link is already in storage")
var Err410 = errors.New("link is deleted, sorry :(")
var Err404 = errors.New("not found")

type Storage interface {
	CreateShort(userID string, urls ...string) ([]string, error)
	GetLong(id string) (string, error)
	Delete(userID string, ids ...string) error
	GetHistory(userID string) ([]LinkJSON, error)
	Ping() error
	GetConfig() config.Config
}

func NewStorage(cfg config.Config) (Storage, error) {
	if cfg.StoragePath != "" {
		return NewFileStorage(cfg)
	}

	if cfg.BasePath != "" {
		return NewDBStorage(cfg)
	}

	return NewMapStorage(cfg)
}

type LinkJSON struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

type BatchJSON struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	URL           string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}
