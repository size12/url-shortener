package handlers

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
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
		gzreader, err := gzip.NewReader(r.Body)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		output, err := ioutil.ReadAll(gzreader)
		r.Body.Close()
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		r.Body = io.NopCloser(strings.NewReader(string(output)))

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
