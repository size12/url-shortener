package linkhelpers

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
)

type URLLinks struct {
	Locations map[string]string
	sync.Mutex
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
