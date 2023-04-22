package handlers

import (
	"compress/gzip"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/size12/url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCookieMiddleware(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	// middleware should add cookie to empty request.
	validCookie := ""

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Cookie("userID")
		assert.NoError(t, ok)
		assert.NotEmpty(t, userID.Value)
		validCookie = userID.Value
	})

	CookieMiddleware(next).ServeHTTP(w, request)

	// middleware should change bad cookie to valid one.

	badCookie := "badCookie12"

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "userID", Value: badCookie, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)

	next = func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Cookie("userID")
		assert.NoError(t, ok)
		assert.NotEqual(t, badCookie, userID.Value)
	}

	// middleware shouldn't change valid cookie.

	cookie = http.Cookie{Name: "userID", Value: validCookie, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)

	next = func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Cookie("userID")
		assert.NoError(t, ok)
		assert.Equal(t, validCookie, userID.Value)
	}
}

func TestNewIPPermissionsChecker(t *testing.T) {
	cfg := config.GetTestConfig()
	request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Error(t, errors.New("shouldn't process this request"))
	})

	NewIPPermissionsChecker(cfg)(next).ServeHTTP(w, request)

	cfg.TrustedSubnet = "127.0.0.1/24"

	request = httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	request.Header.Set("X-Real-IP", "128.0.0.2")
	w = httptest.NewRecorder()

	next = func(w http.ResponseWriter, r *http.Request) {
		assert.Error(t, errors.New("shouldn't process this request"))
	}

	NewIPPermissionsChecker(cfg)(next).ServeHTTP(w, request)

	request = httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	request.Header.Set("X-Real-IP", "127.0.0.1")
	w = httptest.NewRecorder()

	nextWasCalled := false

	next = func(w http.ResponseWriter, r *http.Request) {
		nextWasCalled = true
	}

	NewIPPermissionsChecker(cfg)(next).ServeHTTP(w, request)

	assert.Equal(t, true, nextWasCalled)
}

func TestGzipHandle(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header().Get("Accept-Encoding")
		assert.Equal(t, "", header)
	})

	GzipHandle(next).ServeHTTP(w, request)

	w = httptest.NewRecorder()
	request.Header.Set("Accept-Encoding", "gzip")

	next = func(w http.ResponseWriter, r *http.Request) {
		header := w.Header().Get("Content-Encoding")
		assert.Equal(t, "gzip", header)
	}

	GzipHandle(next).ServeHTTP(w, request)

}

func TestGzipRequest(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	request.Header.Set("Content-Encoding", "gzip")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.IsType(t, &gzip.Reader{}, r.Body)
	})

	GzipRequest(next).ServeHTTP(w, request)

}
