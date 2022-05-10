package handlers

import (
	"bytes"
	"compress/gzip"

	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/SavchenkoOleg/shot/internal/storage"
	"github.com/jackc/pgerrcode"
)

type compressBodyWr struct {
	http.ResponseWriter
	writer io.Writer
}

func (gz compressBodyWr) Write(b []byte) (int, error) {
	return gz.writer.Write(b)
}

func genereteUUIDString(b []byte) string {

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

}

func CookieMiddleware(conf *storage.AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			const CookieUserIDName = "UserIDName"
			var cypher = []byte("xDFaLoYSqcRaHZxs")
			var sk = "verySecretKey"

			cookieUserID, _ := r.Cookie(CookieUserIDName)

			if cookieUserID != nil {

				userIDdata, _ := hex.DecodeString(cookieUserID.Value)

				h := hmac.New(sha256.New, cypher)
				h.Write([]byte(sk))
				sign := h.Sum(nil)

				if bytes.Contains(userIDdata, sign) {
					b := bytes.Replace(userIDdata, sign, []byte(nil), -1)
					conf.UserID = genereteUUIDString(b)

					next.ServeHTTP(w, r)
					return
				}
			}

			userID, _ := generateUUID()
			conf.UserID = genereteUUIDString(userID)

			h := hmac.New(sha256.New, cypher)
			h.Write([]byte(sk))
			CookieUserIDValue := hex.EncodeToString(append(userID, h.Sum(nil)...))
			cookie := http.Cookie{
				Name:   CookieUserIDName,
				Value:  CookieUserIDValue,
				MaxAge: 3600}

			http.SetCookie(w, &cookie)
			next.ServeHTTP(w, r)

		})
	}
}

func HandlerShotJSON(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type inSt struct {
			URL string `json:"url"`
		}
		type outSt struct {
			Result string `json:"result"`
		}

		var resultStatusSucsess int

		resultStatusSucsess = 201

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

		if err != nil && storage.ErrorCode(err) == pgerrcode.UniqueViolation {

			resultStatusSucsess = 409

		} else if err != nil {
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

		w.WriteHeader(resultStatusSucsess)
		w.Write(tx)
	}
}

func HandlerShot(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var resultStatusSucsess int
		resultStatusSucsess = 201

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

		if err != nil && storage.ErrorCode(err) == pgerrcode.UniqueViolation {

			resultStatusSucsess = 409

		} else if err != nil {
			http.Error(w, "internal err", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(resultStatusSucsess)
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

		longURL, exp := storage.RestoreURL(idPath, conf)

		if !exp {
			http.Error(w, "URL for the specified id was not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func HandlerUsershortingList(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		jsonText, err := storage.AllUserActon(conf)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if jsonText == "" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonText))
	}
}

func generateUUID() ([]byte, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	return b, nil
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

func HandlerPingDB(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		dbexp := storage.PingDB(conf)
		if dbexp {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Connect string: " + conf.PgxConnect.Config().ConnString()))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No connection to the Postgres server"))

	}
}

func HandlerShotBach(conf *storage.AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if conf.ConnectionStringDB == "" {
			http.Error(w, "sql connection is not established", http.StatusInternalServerError)
			return
		}

		bodyIn := []storage.ShortenBatchIn{}

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

		bodyOut, err := storage.DBshortenrBatch(bodyIn, conf)

		if err != nil {
			http.Error(w, "internal err", http.StatusInternalServerError)
			return
		}

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
