package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var mapURL = make(map[string]string)
var mapID = make(map[string]string)

type ServConfig struct {
	NewURLPref   string
	ServerAdress string
}

func HendlerSetting() (outConf ServConfig) {

	var newURLPref string
	var serverAdress string
	var exp bool

	outConf.NewURLPref = "newURL"
	outConf.ServerAdress = ":8080"

	newURLPref, exp = os.LookupEnv("BASE_URL")
	if exp {
		outConf.NewURLPref = newURLPref
	}

	serverAdress, exp = os.LookupEnv("SERVER_ADDRESS")
	if exp {
		outConf.ServerAdress = serverAdress
	}

	return outConf
}

func HandlerShotJSON(w http.ResponseWriter, r *http.Request) {

	type inSt struct {
		URL string `json:"url"`
	}
	type outSt struct {
		Result string `json:"result"`
	}

	var NewID string
	var config ServConfig

	config = HendlerSetting()

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

	NewID, exp := mapURL[bodyIn.URL]
	if !exp {
		NewID = "/" + config.NewURLPref + strconv.Itoa(len(mapURL)+1)
		mapURL[bodyURL] = NewID
		mapID[NewID] = bodyURL
	}

	bodyOut.Result = "http://localhost" + config.ServerAdress + NewID

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

	var NewID string
	var config ServConfig

	config = HendlerSetting()

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
		NewID = "/" + config.NewURLPref + strconv.Itoa(len(mapURL)+1)
		mapURL[bodyURL] = NewID
		mapID[NewID] = bodyURL
	}
	NewURL := "http://localhost" + config.ServerAdress + NewID
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
