// Package handlers gets handlers for service endpoints.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/config"
	"github.com/size12/url-shortener/internal/storage"
)

type Handlers interface {
}

type Service struct {
	cfg     config.Config
	storage storage.Storage
}

// NewService gets new handlers service.
func NewService(cfg config.Config, s storage.Storage) *Service {
	return &Service{
		cfg:     cfg,
		storage: s,
	}
}

// CheckPing checks if storage works.
func (service *Service) CheckPing() error {
	return service.storage.Ping()
}

// PingHandler checks if storage works.
func PingHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := service.CheckPing()
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

// DeleteURL deletes link from storage.
// You can delete link, only if you've created it.
func (service *Service) DeleteURL(userID string, urls []string) error {
	return service.storage.Delete(userID, urls...)
}

// DeleteHandler deletes link from storage.
func DeleteHandler(service *Service) http.HandlerFunc {
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

		err = service.DeleteURL(userID, toDelete)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// ShortURLs shorts many urls.
func (service *Service) ShortURLs(userID string, urlsJSON []storage.BatchJSON) ([]storage.BatchJSON, error) {
	urls := make([]string, len(urlsJSON))
	resultJSON := make([]storage.BatchJSON, len(urlsJSON))

	for i := range urlsJSON {
		urls[i] = urlsJSON[i].URL
	}

	result, err := service.storage.CreateShort(userID, urls...)
	if err != nil && err != storage.Err409 {
		return nil, err
	}

	for i, v := range urlsJSON {
		resultJSON[i] = storage.BatchJSON{CorrelationID: v.CorrelationID, ShortURL: service.cfg.BaseURL + "/" + result[i]}
	}

	return resultJSON, err
}

// URLBatchHandler shortens batch of urls in single request.
func URLBatchHandler(service *Service) http.HandlerFunc {
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

		err = json.Unmarshal(resBody, &reqURLs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		respURLs, err := service.ShortURLs(userID, reqURLs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

// GetHistory gets history of your urls.
func (service *Service) GetHistory(userID string) ([]storage.LinkJSON, error) {
	return service.storage.GetHistory(userID)
}

// URLHistoryHandler gets history of your urls.
func URLHistoryHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userID := userCookie.Value

		history, err := service.GetHistory(userID)
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

// GetLongURL gets long url.
func (service *Service) GetLongURL(id string) (string, error) {
	return service.storage.GetLong(id)
}

// URLGetHandler sends person to page, which url was shortened.
func URLGetHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id parameter", http.StatusBadRequest)
			return
		}
		url, err := service.GetLongURL(id)

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

// ShortSingleURL shorts single url.
func (service *Service) ShortSingleURL(userID string, url string) (string, error) {
	result, err := service.storage.CreateShort(userID, url)
	if len(result) == 0 {
		return "", err
	}
	return result[0], err
}

// URLPostHandler creates new short URL.
func URLPostHandler(service *Service) http.HandlerFunc {
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
				res, err2 := service.ShortSingleURL(userID, reqJSON.URL)

				if err2 != nil && !errors.Is(err2, storage.Err409) {
					http.Error(w, err2.Error(), http.StatusBadRequest)
					return
				}

				respJSON, err := json.Marshal(storage.ResponseJSON{Result: service.cfg.BaseURL + "/" + res})

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
				res, err2 := service.ShortSingleURL(userID, string(resBody))
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
				w.Write([]byte(service.cfg.BaseURL + "/" + res))
			}

		}

	}
}

// GetStatistic returns total urls and users.
func (service *Service) GetStatistic() (storage.Statistic, error) {
	return service.storage.GetStatistic()
}

// StatisticHandler returns total urls and users.
func StatisticHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := service.GetStatistic()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(stats)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
