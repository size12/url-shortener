package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestURLHandler(t *testing.T) {
	var code int = 201
	request, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
	w := httptest.NewRecorder()
	h := http.HandlerFunc(URLHandler)
	h.ServeHTTP(w, request)
	res := w.Result()
	if res.StatusCode != code {
		t.Error("Wrong code")
	}
}
