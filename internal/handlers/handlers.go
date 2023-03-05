// Package handlers gets handlers for service endpoints.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/storage"
)

// PingHandler checks if storage works.
func PingHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.Ping()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

// URLErrorHandler returns error, if someone gets wrong page.
func URLErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "wrong method", http.StatusBadRequest)
}

// DeleteHandler deletes link from storage.
// You can delete link, only if you've created it.
func DeleteHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userID := userCookie.Value

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", http.StatusBadRequest)
			return
		}

		var toDelete []string
		err = json.Unmarshal(resBody, &toDelete)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.Delete(userID, toDelete...)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// URLBatchHandler shortens batch of urls in single request.
func URLBatchHandler(s storage.Storage) http.HandlerFunc {
	cfg := s.GetConfig()
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userID := userCookie.Value

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", http.StatusBadRequest)
			return
		}

		var reqURLs []storage.BatchJSON
		var respURLs []storage.BatchJSON

		err = json.Unmarshal(resBody, &reqURLs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var urls []string

		for _, v := range reqURLs {
			urls = append(urls, v.URL)
		}

		res, err := s.CreateShort(userID, urls...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for i, v := range reqURLs {
			respURLs = append(respURLs, storage.BatchJSON{CorrelationID: v.CorrelationID, ShortURL: cfg.BaseURL + "/" + res[i]})
		}

		b, err := json.Marshal(respURLs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(b)
	}
}

// URLHistoryHandler gets history of your urls.
func URLHistoryHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		userID := userCookie.Value

		history, err := s.GetHistory(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if len(history) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		data, err := json.Marshal(history)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	}
}

// URLGetHandler sends person to page, which url was shortened.
func URLGetHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id parameter", http.StatusBadRequest)
			return
		}
		url, err := s.GetLong(id)

		if errors.Is(err, storage.Err410) {
			http.Error(w, "link is deleted", http.StatusGone)
			return
		}

		if errors.Is(err, storage.Err404) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// URLPostHandler creates new short URL.
func URLPostHandler(s storage.Storage) http.HandlerFunc {
	cfg := s.GetConfig()
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		userID := userCookie.Value
		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", http.StatusBadRequest)
			return
		}
		switch r.Header.Get("Content-Type") {
		case "application/json":
			{
				var reqJSON storage.RequestJSON
				err := json.Unmarshal(resBody, &reqJSON)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				res, err2 := s.CreateShort(userID, reqJSON.URL)

				if err2 != nil && !errors.Is(err2, storage.Err409) {
					http.Error(w, err2.Error(), http.StatusBadRequest)
					return
				}

				respJSON, err := json.Marshal(storage.ResponseJSON{Result: cfg.BaseURL + "/" + res[0]})

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				if errors.Is(err2, storage.Err409) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusCreated)
				}
				w.Write(respJSON)
			}
		default:
			{
				res, err2 := s.CreateShort(userID, string(resBody))
				if err2 != nil && !errors.Is(err2, storage.Err409) {
					http.Error(w, err2.Error(), 400)
					return
				}
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				if errors.Is(err2, storage.Err409) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusCreated)
				}
				w.Write([]byte(cfg.BaseURL + "/" + res[0]))
			}

		}

	}
}
