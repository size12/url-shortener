package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"io"
	"net/http"
	"strings"
)

func URLErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "wrong method", 400)
}

func URLGetHandler(cfg config.Config, links linkhelpers.Storage) http.HandlerFunc {
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

func URLPostHandler(cfg config.Config, links linkhelpers.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", 400)
			return
		}

		switch r.Header.Get("Content-Type") {
		case "application/json":
			{
				var reqJSON linkhelpers.RequestJSON
				err := json.Unmarshal(resBody, &reqJSON)
				if err != nil {
					http.Error(w, err.Error(), 400)
				}
				res, err := links.NewShortURL(reqJSON.URL)

				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				respJSON, err := json.Marshal(linkhelpers.ResponseJSON{Result: cfg.BaseURL + "/" + res})

				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(w.Header().Get("Content-Type"), "gzip") {
					w.Header().Add("Content-Type", http.DetectContentType(respJSON))
				}
				w.WriteHeader(201)
				w.Write(respJSON)
			}
		default:
			{
				res, err := links.NewShortURL(string(resBody))
				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				if strings.Contains(w.Header().Get("Content-Type"), "gzip") {
					w.Header().Add("Content-Type", "gzip")
				}
				w.WriteHeader(201)
				w.Write([]byte(cfg.BaseURL + "/" + res))
			}

		}

	}
}
