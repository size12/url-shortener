package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
