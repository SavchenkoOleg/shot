package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/SavchenkoOleg/shot/internal/storage"
)

func HandlerShotJSON(w http.ResponseWriter, r *http.Request) {

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

	resultURL, err := storage.ReductionURL(bodyIn.URL)

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

func HandlerShot(w http.ResponseWriter, r *http.Request) {

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

	shotURL, err := storage.ReductionURL(longURL)

	if err != nil {
		http.Error(w, "internal err", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
	w.Write([]byte(shotURL))
}

func HandlerIndex(w http.ResponseWriter, r *http.Request) {

	idPath := r.URL.Path
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
