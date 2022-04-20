package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/SavchenkoOleg/shot/internal/storage"
)

type compressBodyWr struct {
	http.ResponseWriter
	writer io.Writer
}

func (gz compressBodyWr) Write(b []byte) (int, error) {
	return gz.writer.Write(b)
}

func CompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()

		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		next.ServeHTTP(compressBodyWr{
			ResponseWriter: w,
			writer:         gz,
		}, r)
	})
}

func HandlerShotJSON(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type inSt struct {
			URL string `json:"url"`
		}
		type outSt struct {
			Result string `json:"result"`
		}

		bodyIn := inSt{}
		bodyOut := outSt{}

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := json.Unmarshal(b, &bodyIn); err != nil {
			http.Error(w, "uncorrect body URL format", http.StatusBadRequest)
			return
		}

		bodyURL := bodyIn.URL
		if bodyURL == "" {
			http.Error(w, "uncorrect body URL format", http.StatusBadRequest)
			return
		}

		resultURL, err := storage.ReductionURL(bodyIn.URL, conf)

		if err != nil {
			http.Error(w, "internal err", http.StatusInternalServerError)
			return
		}

		bodyOut.Result = resultURL

		tx, err := json.Marshal(bodyOut)

		if err != nil {
			http.Error(w, "internal err", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		w.WriteHeader(201)
		w.Write(tx)
	}
}

func HandlerShot(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		longURL := string(b)

		if len(longURL) == 0 {
			http.Error(w, "uncorrect reduction URL", http.StatusBadRequest)
			return
		}

		shotURL, err := storage.ReductionURL(longURL, conf)

		if err != nil {
			http.Error(w, "internal err", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(201)
		w.Write([]byte(shotURL))
	}
}

func HandlerIndex(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idPath := strings.Replace(r.URL.Path, "/"+conf.BaseURL+"/", "", 1)

		if idPath == "" {
			http.Error(w, "The parameter is missing", http.StatusBadRequest)
			return
		}

		longURL, exp := storage.RestoreURL(idPath)

		if !exp {
			http.Error(w, "URL for the specified id was not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
