package link_helpers

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
)

type UrlLinks struct {
	Locations map[string]string
	sync.Mutex
}

var Links UrlLinks

func NewShortUrl(longUrl string) (string, error) {
	if Links.Locations == nil {
		Links.Locations = make(map[string]string)
	}

	if _, err := url.ParseRequestURI(longUrl); err != nil {
		return "", errors.New("Wrong link")
	}
	Links.Lock()
	defer Links.Unlock()
	lastId := len(Links.Locations)
	newId := fmt.Sprint(lastId + 1)
	Links.Locations[newId] = longUrl
	return fmt.Sprint(lastId + 1), nil
}

func GetFullUrl(id string) (string, error) {
	if Links.Locations == nil {
		Links.Locations = make(map[string]string)
	}

	Links.Lock()
	defer Links.Unlock()
	if el, ok := Links.Locations[id]; ok {
		return el, nil
	}
	return "", errors.New("No such id")
}
