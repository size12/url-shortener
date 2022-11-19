package handlers

import (
	"github.com/size12/url-shortener/internal/linkhelpers"
	"io"
	"net/http"
)

func URLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "wrong method", 400)
		return
	}

	if r.URL.Path != "/" {
		http.Error(w, "wrong url", 400)
		return
	}

	if r.Method == http.MethodGet {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id parameter", 400)
			return
		}
		url, err := linkhelpers.GetFullURL(id)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}

	if r.Method == http.MethodPost {
		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", 400)
			return
		}

		res, err := linkhelpers.NewShortURL(string(resBody))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.WriteHeader(201)
		w.Write([]byte(res))
		return
	}
}
