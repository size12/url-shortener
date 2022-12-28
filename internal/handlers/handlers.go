package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"io"
	"net/http"
	"time"
)

func PingHandler(links linkhelpers.URLLinks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if links.DB != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			if err := links.DB.PingContext(ctx); err != nil {
				http.Error(w, "DataBase is not working", 500)
				return
			}
			w.WriteHeader(200)
		} else {
			http.Error(w, "DataBase is not working", 500)
			return
		}
	}
}

func URLErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "wrong method", 400)
}

func URLBatchHandler(links linkhelpers.URLLinks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		userID := userCookie.Value
		_ = userID

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", 400)
			return
		}

		var reqURLs []linkhelpers.BatchJSON
		var respURLs []linkhelpers.BatchJSON

		err = json.Unmarshal(resBody, &reqURLs)

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		var urls []string

		for _, v := range reqURLs {
			urls = append(urls, v.URL)
		}

		res, err := links.NewShortURL(userID, urls...)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		for i, v := range reqURLs {
			respURLs = append(respURLs, linkhelpers.BatchJSON{CorrelationID: v.CorrelationID, ShortURL: links.Cfg.BaseURL + "/" + res[i]})
		}

		b, err := json.Marshal(respURLs)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(b)
	}
}

func URLHistoryHandler(links linkhelpers.URLLinks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		userID := userCookie.Value

		historyShort := links.Users[userID]
		var history []linkhelpers.LinkJSON

		for _, short := range historyShort {
			long, err := links.GetFullURL(short)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			history = append(history, linkhelpers.LinkJSON{ShortURL: links.Cfg.BaseURL + "/" + short, LongURL: long})
		}

		if len(history) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		data, err := json.Marshal(history)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	}
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
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		userID := userCookie.Value
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
					return
				}
				res, err2 := links.NewShortURL(userID, reqJSON.URL)

				if err2 != nil && !errors.As(err2, &linkhelpers.Error409) {
					http.Error(w, err2.Error(), 400)
					return
				}

				respJSON, err := json.Marshal(linkhelpers.ResponseJSON{Result: links.Cfg.BaseURL + "/" + res[0]})

				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				if errors.As(err2, &linkhelpers.Error409) {
					w.WriteHeader(409)
				} else {
					w.WriteHeader(201)
				}
				w.Write(respJSON)
			}
		default:
			{
				res, err2 := links.NewShortURL(userID, string(resBody))
				if err2 != nil && !errors.As(err2, &linkhelpers.Error409) {
					http.Error(w, err2.Error(), 400)
					return
				}
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				if errors.As(err2, &linkhelpers.Error409) {
					w.WriteHeader(409)
				} else {
					w.WriteHeader(201)
				}
				w.Write([]byte(links.Cfg.BaseURL + "/" + res[0]))
			}

		}

	}
}
