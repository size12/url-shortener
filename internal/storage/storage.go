// Package storage creates storage and stores data in it.
package storage

import (
	"errors"

	"github.com/size12/url-shortener/internal/config"
)

// Errors for storage response.
var (
	Err409 = errors.New("link is already in storage")
	Err410 = errors.New("link is deleted, sorry :(")
	Err404 = errors.New("not found")
)

// Storage is an interface that describes storage.
type Storage interface {
	CreateShort(userID string, urls ...string) ([]string, error)
	GetLong(id string) (string, error)
	Delete(userID string, ids ...string) error
	GetHistory(userID string) ([]LinkJSON, error)
	Ping() error
	GetConfig() config.Config
	GetStatistic() (Statistic, error)
}

// NewStorage creates new storage based on config.
func NewStorage(cfg config.Config) (Storage, error) {
	if cfg.BasePath != "" {
		return NewDBStorage(cfg)
	}

	if cfg.StoragePath != "" {
		return NewFileStorage(cfg)
	}

	return NewMapStorage(cfg)
}

// Structs for response.

// Statistic struct for statistic which contains total shortened URLs number and users number.
type Statistic struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}

// LinkJSON struct for history response.
type LinkJSON struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

// BatchJSON struct for batch request.
type BatchJSON struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	URL           string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}

// RequestJSON struct for single application/json request.
type RequestJSON struct {
	URL string `json:"url"`
}

// ResponseJSON struct for single application/json response.
type ResponseJSON struct {
	Result string `json:"result"`
}
