package linkhelpers

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/size12/url-shortener/internal/config"
	"net/url"
	"os"
	"sync"
)

type Storage interface {
	NewShortURL(longURL string) (string, error)
	GetFullURL(id string) (string, error)
}

type URLLinks struct {
	Locations map[string]string
	*sync.Mutex
	File *os.File
}

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

func NewStorage(cfg config.Config) (Storage, error) {
	loc := make(map[string]string)
	if cfg.StoragePath != "" {
		file, err := os.OpenFile(cfg.StoragePath+"/saved.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0777)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		i := 1
		for scanner.Scan() {
			loc[fmt.Sprint(i)] = scanner.Text()
			i += 1
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return &URLLinks{Locations: loc, Mutex: &sync.Mutex{}, File: file}, nil
	}
	return &URLLinks{Locations: loc, Mutex: &sync.Mutex{}}, nil
}

func (Links *URLLinks) NewShortURL(longURL string) (string, error) {
	if _, err := url.ParseRequestURI(longURL); err != nil {
		return "", errors.New("wrong link") //checks if url valid
	}
	Links.Lock()
	defer Links.Unlock()
	lastID := len(Links.Locations)
	newID := fmt.Sprint(lastID + 1)
	Links.Locations[newID] = longURL
	fmt.Println(Links.File)
	if Links.File != nil {
		_, err := Links.File.Write([]byte(longURL + "\n"))
		Links.File.Sync()
		if err != nil {
			return "", err
		}
	}
	fmt.Println(Links)
	return newID, nil
}

func (Links *URLLinks) GetFullURL(id string) (string, error) {
	Links.Lock()
	defer Links.Unlock()
	if el, ok := Links.Locations[id]; ok {
		return el, nil
	}
	return "", errors.New("no such id")
}
