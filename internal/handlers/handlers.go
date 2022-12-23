package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"io"
	"net/http"
)

func URLErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "wrong method", 400)
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
				res, err := links.NewShortURL(reqJSON.URL)

				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				links.Users[userID] = append(links.Users[userID], res)
				respJSON, err := json.Marshal(linkhelpers.ResponseJSON{Result: links.Cfg.BaseURL + "/" + res})

				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				w.Header().Set("Content-Type", "application/json")
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
				links.Users[userID] = append(links.Users[userID], res)
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(201)
				w.Write([]byte(links.Cfg.BaseURL + "/" + res))
			}

		}

	}
}
