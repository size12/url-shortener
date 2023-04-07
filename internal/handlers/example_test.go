package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/storage"
)

func ExampleURLPostHandler() {
	// data.
	body := "https://yandex.ru"
	// Generating handler.
	cfg := config.Config{BaseURL: "http://127.0.0.1:8080"}

	s, err := storage.NewMapStorage(cfg)

	if err != nil {
		log.Fatal("Failed get storage")
	}

	h := URLPostHandler(s)

	// Generating request.
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookieString := "user12cookie"
	cookie := http.Cookie{Name: "userID", Value: cookieString, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, request)

	// Checking result.
	res := w.Result()

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	fmt.Printf("Short URL is: %s\n", resBody)
	fmt.Printf("Now in storage: %s\n", s.Locations)
	// Output:
	// Short URL is: http://127.0.0.1:8080/1
	// Now in storage: map[1:https://yandex.ru]
}

func ExampleURLBatchHandler() {
	// data.
	body := `[{"correlation_id": "1", "original_url": "https://yandex.ru"}, 
			  {"correlation_id": "2", "original_url": "https://google.com"}, 
			  {"correlation_id": "3", "original_url": "https://youtube.com"}]`
	// Generating handler.
	cfg := config.Config{BaseURL: "http://127.0.0.1:8080"}

	s, err := storage.NewMapStorage(cfg)

	if err != nil {
		log.Fatal("Failed get storage")
	}

	h := URLBatchHandler(s)

	// Generating request.
	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookieString := "user12cookie"
	cookie := http.Cookie{Name: "userID", Value: cookieString, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, request)

	// Checking result.
	res := w.Result()

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	fmt.Printf("Short URL is: %s\n", resBody)
	fmt.Printf("Now in storage: %s\n", s.Locations)
	// Output:
	// Short URL is: [{"correlation_id":"1","short_url":"http://127.0.0.1:8080/1"},{"correlation_id":"2","short_url":"http://127.0.0.1:8080/2"},{"correlation_id":"3","short_url":"http://127.0.0.1:8080/3"}]
	// Now in storage: map[1:https://yandex.ru 2:https://google.com 3:https://youtube.com]
}

func ExampleURLHistoryHandler() {
	// data.
	userID := "user12cookie"
	// generating storage.
	cfg := config.Config{BaseURL: "http://127.0.0.1:8080"}

	s, err := storage.NewMapStorage(cfg)

	if err != nil {
		log.Fatal("Failed get storage")
	}

	// creating short urls.
	_, err = s.CreateShort(userID, "https://yandex.ru")

	if err != nil {
		log.Fatal("Failed shorten URL")
	}

	_, err = s.CreateShort(userID, "https://google.com")

	if err != nil {
		log.Fatal("Failed shorten URL")
	}
	// Generating handler.
	h := URLHistoryHandler(s)

	// Generating request.
	request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "userID", Value: userID, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, request)

	// Checking result.
	res := w.Result()

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	fmt.Printf("URLs history is: %s\n", resBody)
	// Output:
	// URLs history is: [{"short_url":"http://127.0.0.1:8080/1","original_url":"https://yandex.ru"},{"short_url":"http://127.0.0.1:8080/2","original_url":"https://google.com"}]
}

func ExampleDeleteHandler() {
	// data.
	data := `["1"]`
	userID := "user12cookie"
	// generating storage.
	cfg := config.Config{BaseURL: "http://127.0.0.1:8080"}

	s, err := storage.NewMapStorage(cfg)

	if err != nil {
		log.Fatal("Failed get storage")
	}

	// creating short urls.
	_, err = s.CreateShort(userID, "https://yandex.ru")

	if err != nil {
		log.Fatal("Failed shorten URL")
	}

	_, err = s.CreateShort(userID, "https://google.com")

	if err != nil {
		log.Fatal("Failed shorten URL")
	}

	// Generating handler.
	h := DeleteHandler(s)

	// Generating request.
	request := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(data))
	w := httptest.NewRecorder()

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "userID", Value: userID, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, request)

	// Checking result.

	fmt.Printf("URLs in storage are: %v\n", s.Locations)
	fmt.Printf("Links which deleted: %v\n", s.Deleted)

	// Output:
	// URLs in storage are: map[1:https://yandex.ru 2:https://google.com]
	// Links which deleted: map[1:true]
}
