package handlers

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

//compress response

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := r.Cookie("userID")
		_ = userID
		if err != nil {
			randomID, err := generateRandom(8)
			if err != nil {
				http.Error(w, err.Error(), 400)
			}
			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookieString := hex.EncodeToString(randomID)
			fmt.Println("New user cookie is:", cookieString)
			cookie := http.Cookie{Name: "userID", Value: cookieString, Expires: expiration, Path: "/"}
			http.SetCookie(w, &cookie)
			r.AddCookie(&cookie)
		}
		next.ServeHTTP(w, r)
	})
}

func GzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// переменная r будет читать входящие данные и распаковывать их
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer reader.Close()
		r.Body = reader
		next.ServeHTTP(w, r)
	})
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
