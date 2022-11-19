package handlers

import (
	"fmt"
	"github.com/size12/url-shortener/internal/linkhelpers"
	"io"
	"net/http"
	"strings"
)

func URLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "wrong method", 400)
		return
	}

	if r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/")
		if id == "" {
			http.Error(w, "missing id parameter", 400)
			return
		}
		url, err := linkhelpers.GetFullURL(id)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		fmt.Println("Redirecting to:", url)
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		fmt.Println(w)
		//w.Write([]byte(""))
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

		fmt.Println("New id:", res, string(resBody))
		w.WriteHeader(201)
		w.Write([]byte(res))
		return
	}
}
