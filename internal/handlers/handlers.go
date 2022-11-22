package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"io"
	"net/http"
)

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "wrong method", 400)
}

func URLGetHandler(links linkhelpers.URLLinks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id parameter", 400)
			return
		}
		url, err := links.GetFullURL(id)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func URLPostHandler(links linkhelpers.URLLinks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", 400)
			return
		}

		res, err := links.NewShortURL(string(resBody))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(201)
		w.Write([]byte("http://127.0.0.1:8080/" + res))
	}
}
