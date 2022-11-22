package handlers

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestURLErrorHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodDelete, "/1", nil)
	w := httptest.NewRecorder()
	h := http.HandlerFunc(URLErrorHandler)
	h.ServeHTTP(w, request)
	assert.Equal(t, 400, w.Result().StatusCode)
	w.Result().Body.Close()
}

func TestURLPostHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		links    linkhelpers.URLLinks
		error    bool
	}
	cases := []struct {
		name  string
		links linkhelpers.URLLinks
		url   string
		want  want
	}{
		{
			"add new link storage",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"https://google.com",
			want{201, "2", linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru", "2": "https://google.com"}}, false},
		},
		{
			"add bad link to storage",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"efjwejfekw",
			want{400, "wrong link\n", linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}}, true},
		},
		{
			"don't send body",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"",
			want{400, "wrong body\n", linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}}, true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.url))
			w := httptest.NewRecorder()
			h := URLPostHandler(tc.links)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.NoError(t, err)
			if tc.want.error {
				assert.Equal(t, tc.want.response, string(resBody))
			}
			assert.Equal(t, tc.want.links, tc.links)

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
		links linkhelpers.URLLinks
		id    string
		want  want
	}{
		{
			"get link which in storage",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "http://dzen.ru"}},
			"1",
			want{307, "", false},
		},
		{
			"get link which NOT in storage",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"2",
			want{400, "no such id\n", true},
		},
		{
			"don't send ID parameter",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
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

func TestNewShortURL(t *testing.T) {
	type want struct {
		links linkhelpers.URLLinks
		id    string
		error error
	}
	cases := []struct {
		name  string
		links linkhelpers.URLLinks
		url   string
		want  want
	}{
		{
			"add new link",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"https://google.com",
			want{
				linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru", "2": "https://google.com"}},
				"2",
				nil,
			},
		},
		{
			"add bad link",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"njkjnekjre",
			want{
				linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
				"",
				errors.New("wrong link"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tc.links.NewShortURL(tc.url)
			assert.Equal(t, tc.want.links, tc.links)
			if tc.want.error != nil {
				assert.Equal(t, tc.want.error, err)
			} else {
				assert.Equal(t, tc.want.id, id)
			}
		})
	}
}

func TestGetFullURL(t *testing.T) {
	type want struct {
		url   string
		error error
	}
	cases := []struct {
		name  string
		links linkhelpers.URLLinks
		id    string
		want  want
	}{
		{
			"get existed link",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"1",
			want{
				"https://dzen.ru",
				nil,
			},
		},
		{
			"get non-existed link",
			linkhelpers.URLLinks{Locations: map[string]string{"1": "https://dzen.ru"}},
			"2",
			want{
				"",
				errors.New("no such id"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := tc.links.GetFullURL(tc.id)
			if tc.want.error != nil {
				assert.Equal(t, tc.want.error, err)
			} else {
				assert.Equal(t, tc.want.url, url)
			}
		})
	}
}
