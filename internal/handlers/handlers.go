package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

var mapURL = make(map[string]string)
var mapID = make(map[string]string)
var newURLPref = "newURL"
var localURL = "http://localhost:8080/"

func HandlerShot(w http.ResponseWriter, r *http.Request) {

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

func HandlerIndex(w http.ResponseWriter, r *http.Request) {

	idPath := r.URL.Path
	if idPath == "" {
		http.Error(w, "The parameter is missing", http.StatusBadRequest)
		return
	}
	id := strings.TrimPrefix(idPath, "/")
	url, exp := mapID[id]

	if !exp {
		http.Error(w, "URL for the specified id was not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)

}
