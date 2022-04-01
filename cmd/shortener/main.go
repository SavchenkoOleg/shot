package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

var mapURL = make(map[string]string)
var mapID = make(map[string]string)
var newURLPref = "newURL"
var localURL = "http://localhost:8080/"

func handlerShot(w http.ResponseWriter, r *http.Request) {

	var NewID string

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	bodyURL := string(b)

	if len(bodyURL) == 0 {
		http.Error(w, "uncorrect reduction URL", http.StatusBadRequest)
		return
	}

	NewID, exp := mapURL[bodyURL]
	if !exp {
		NewID = newURLPref + strconv.Itoa(len(mapURL)+1)
		mapURL[bodyURL] = NewID
		mapID[NewID] = bodyURL
	}
	NewURL := localURL + NewID
	w.WriteHeader(201)
	w.Write([]byte(NewURL))
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "The id parameter is missing", http.StatusBadRequest)
		return
	}
	url, exp := mapID[id]

	if !exp {
		http.Error(w, "URL for the specified id was not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func firstResort(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		handlerShot(w, r)
	case http.MethodGet:
		handlerIndex(w, r)
	default:
		http.Error(w, "Only POST or GET requests are allowed!", http.StatusMethodNotAllowed)
	}

}
func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", firstResort)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}

}
