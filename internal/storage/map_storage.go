package storage

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/size12/url-shortener/internal/config"
)

// MapStorage is storage that storages in map.
type MapStorage struct {
	Cfg       config.Config
	Locations map[string]string
	Users     map[string][]string
	Deleted   map[string]bool
	*sync.Mutex
}

func NewMapStorage(cfg config.Config) (*MapStorage, error) {
	loc := make(map[string]string)
	users := make(map[string][]string)
	deleted := make(map[string]bool)

	return &MapStorage{Locations: loc, Users: users, Deleted: deleted, Cfg: cfg, Mutex: &sync.Mutex{}}, nil
}

// Interface storage.Storage implementation.

func (s *MapStorage) GetConfig() config.Config {
	return s.Cfg
}

func (s *MapStorage) Ping() error {
	return nil
}

func (s *MapStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	result := make([]string, 0)
	s.Lock()
	defer s.Unlock()

	var isErr409 error

	for _, longURL := range urls {
		if _, err := url.ParseRequestURI(longURL); err != nil {
			return nil, errors.New("wrong link " + longURL) //checks if url valid
		}

		lastID := len(s.Locations)
		newID := fmt.Sprint(lastID + 1)
		foundThisLink := false
		for id, link := range s.Locations {
			if link == longURL {
				newID = id
				isErr409 = Err409
				foundThisLink = true
				break
			}
		}

		result = append(result, newID)
		if foundThisLink {
			continue //do not add to storage again
		}

		s.Locations[newID] = longURL
		s.Users[userID] = append(s.Users[userID], newID)
	}

	return result, isErr409
}

func (s *MapStorage) GetLong(id string) (string, error) {
	s.Lock()
	defer s.Unlock()
	if el, ok := s.Locations[id]; ok {
		var isErr410 error
		if s.Deleted[id] {
			isErr410 = Err410
		}
		return el, isErr410
	}
	return "", Err404
}

func (s *MapStorage) Delete(userID string, ids ...string) error {
	s.Lock()
	defer s.Unlock()
	canDelete := s.Users[userID]

	for _, id := range ids {
		for _, can := range canDelete {
			if id == can {
				s.Deleted[id] = true
				break
			}
		}
	}
	return nil
}

func (s *MapStorage) GetHistory(userID string) ([]LinkJSON, error) {
	s.Lock()
	defer s.Unlock()

	historyShort := s.Users[userID]
	var history = make([]LinkJSON, len(historyShort))

	for i, id := range historyShort {
		long := s.Locations[id]
		history[i] = LinkJSON{ShortURL: s.Cfg.BaseURL + "/" + id, LongURL: long}
	}
	return history, nil
}
