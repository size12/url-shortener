package handlers

import (
	link_helpers "github.com/size12/url-shortener/internal/link-helpers"
	"io"
	"net/http"
)

func UrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Wrong method", 400)
		return
	}

	if r.URL.Path != "/" {
		http.Error(w, "Wrong url", 400)
		return
	}

	if r.Method == http.MethodGet {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", 400)
			return
		}
		url, err := link_helpers.GetFullUrl(id)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		http.Redirect(w, r, url, 307)
	}

	if r.Method == http.MethodPost {
		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "Wrong body", 400)
			return
		}

		res, err := link_helpers.NewShortUrl(string(resBody))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.WriteHeader(201)
		w.Write([]byte(res))
		return
	}
}
