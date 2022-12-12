package handlers

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Request    *http.Request
	statusCode int
}

//compress response

func (gw gzipWriter) WriteHeader(statusCode int) {
	gw.statusCode = statusCode
}

func (gw gzipWriter) Write(b []byte) (int, error) {
	if !strings.Contains(gw.Request.Header.Get("Accept-Encoding"), "gzip") {
		return gw.ResponseWriter.Write(b)
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(b)
	if err != nil {
		return 0, errors.New("failed write data to buffer")
	}
	err = w.Close()
	if err != nil {
		return 0, errors.New("failed compress data")
	}
	gw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	gw.ResponseWriter.WriteHeader(gw.statusCode)
	return gw.ResponseWriter.Write(buf.Bytes())
}

func MiddlewareGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gw := gzipWriter{ResponseWriter: w, Request: r, statusCode: 400}
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(gw, r)
			return
		}
		reader, err := gzip.NewReader(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(gw, "can't read body", 400)
		}
		var b bytes.Buffer
		_, err = b.ReadFrom(reader)
		if err != nil {
			http.Error(gw, "something wrong", 400)
		}
		fmt.Println("123")
		r.Body = ioutil.NopCloser(&b)
		next.ServeHTTP(gw, r)
	})
}
