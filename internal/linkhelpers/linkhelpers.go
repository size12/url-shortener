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

var Links URLLinks

func NewShortURL(longURL string) (string, error) {
	if Links.Locations == nil {
		Links.Locations = make(map[string]string)
	}

	if _, err := url.ParseRequestURI(longURL); err != nil {
		return "", errors.New("wrong link")
	}
	Links.Lock()
	defer Links.Unlock()
	lastID := len(Links.Locations)
	newID := fmt.Sprint(lastID + 1)
	Links.Locations[newID] = longURL
	fmt.Println(Links.Locations)
	return newID, nil
}

func GetFullURL(id string) (string, error) {
	if Links.Locations == nil {
		Links.Locations = make(map[string]string)
	}

	Links.Lock()
	defer Links.Unlock()
	if el, ok := Links.Locations[id]; ok {
		return el, nil
	}
	fmt.Println(Links.Locations)
	return "", errors.New("no such id")
}
