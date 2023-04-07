package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestURLErrorHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodDelete, "/1", nil)
	w := httptest.NewRecorder()
	h := http.HandlerFunc(URLErrorHandler)
	h.ServeHTTP(w, request)
	res := w.Result()
	assert.Equal(t, 400, res.StatusCode)
	io.Copy(io.Discard, res.Body)
	defer res.Body.Close()
}

func TestURLPostHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		links    *storage.MapStorage
		error    bool
	}
	cases := []struct {
		name  string
		links *storage.MapStorage
		url   string
		want  want
	}{
		{
			"add new link storage",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{}},
			"https://google.com",
			want{201, "2", &storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru", "2": "https://google.com"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}}, false},
		},
		{
			"add bad link to storage",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}},
			"efjwejfekw",
			want{400, "wrong link", &storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}}, true},
		},
		{
			"don't send body",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}},
			"",
			want{400, "wrong body\n", &storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}}, true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.url))
			w := httptest.NewRecorder()
			cfg := config.GetTestConfig()
			tc.links.Cfg = cfg
			h := URLPostHandler(tc.links)
			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookieString := "123456"
			cookie := http.Cookie{Name: "userID", Value: cookieString, Expires: expiration, Path: "/"}
			request.AddCookie(&cookie)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.NoError(t, err)
			if tc.want.error {
				assert.Contains(t, string(resBody), tc.want.response)
			}
			assert.Equal(t, tc.want.links.Locations, tc.links.Locations)

		})
	}

}

func TestURLPostJSONHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		links    *storage.MapStorage
		error    bool
	}
	cases := []struct {
		name  string
		links *storage.MapStorage
		url   string
		want  want
	}{
		{
			"add new link storage",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{}},
			`{"url":"https://google.com"}`,
			want{201, "2", &storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru", "2": "https://google.com"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}}, false},
		},
		{
			"add bad link to storage",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}},
			"{efjwejfekw",
			want{400, "wrong link\n", &storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}}, true},
		},
		{
			"don't send body",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}},
			"",
			want{400, "wrong body\n", &storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}, Users: map[string][]string{"123456": {"2"}}}, true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tc.url))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			cfg := config.GetTestConfig()
			tc.links.Cfg = cfg
			h := URLPostHandler(tc.links)
			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookieString := "123456"
			cookie := http.Cookie{Name: "userID", Value: cookieString, Expires: expiration, Path: "/"}
			request.AddCookie(&cookie)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			_, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.NoError(t, err)
			assert.Equal(t, tc.want.links.Locations, tc.links.Locations)

		})
	}

}

func TestURLGetHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		error    bool
	}
	cases := []struct {
		name  string
		links *storage.MapStorage
		id    string
		want  want
	}{
		{
			"get link which in storage",
			&storage.MapStorage{Locations: map[string]string{"1": "http://dzen.ru"}, Mutex: &sync.Mutex{}},
			"1",
			want{307, "", false},
		},
		{
			"get link which NOT in storage",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}},
			"2",
			want{404, "not found\n", true},
		},
		{
			"don't send ID parameter",
			&storage.MapStorage{Locations: map[string]string{"1": "https://dzen.ru"}, Mutex: &sync.Mutex{}},
			"",
			want{400, "missing id parameter\n", true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			cfg := config.GetTestConfig()
			tc.links.Cfg = cfg
			h := URLGetHandler(tc.links)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.NoError(t, err)
			if tc.want.error {
				assert.Equal(t, tc.want.response, string(resBody))
			}

		})
	}
}

func TestPingHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	cfg := config.GetTestConfig()
	s, err := storage.NewMapStorage(cfg)
	assert.NoError(t, err)

	h := PingHandler(s)
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookieString := "user12"
	cookie := http.Cookie{Name: "userID", Value: cookieString, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)
	h.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGenerateRandom(t *testing.T) {
	// Generate random bytes.
	_, err := generateRandom(20)
	assert.NoError(t, err, "Generate random bytes.")

	// Generate 0 length random bytes.
	res, err := generateRandom(0)
	assert.NoError(t, err, "Generate 0 length random bytes.")
	assert.Len(t, res, 0, "Generate 0 length random bytes.")
}
