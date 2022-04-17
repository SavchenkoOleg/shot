package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
)

var mapLongURL = make(map[string]string)
var mapShotURL = make(map[string]string)

type ServConfig struct {
	NewURLPref   string
	ServerAdress string
	BaseURL      string
	FullPathTest string
}

func ReductionURL(longURL string) (shotURL string) {

	config := HendlerSetting()

	idURL, exp := mapLongURL[longURL]
	if !exp {
		idURL = "/" + config.NewURLPref + strconv.Itoa(len(mapLongURL)+1)
		mapLongURL[longURL] = idURL
		mapShotURL[idURL] = longURL
	}

	return config.BaseURL + config.ServerAdress + idURL
}

func RestoreURL(shotURL string) (restURL string, exp bool) {

	restURL, exp = mapShotURL[shotURL]

	return restURL, exp
}

func HendlerSetting() (outConf ServConfig) {

	var serverAdress string
	var exp bool

	outConf.NewURLPref = "newURL"
	outConf.ServerAdress = ":8080"
	outConf.BaseURL = "http://localhost"

	BaseURL, exp := os.LookupEnv("BASE_URL")
	if exp {
		outConf.BaseURL = BaseURL
	}

	serverAdress, exp = os.LookupEnv("SERVER_ADDRESS")
	if exp {
		outConf.ServerAdress = serverAdress
	}

	outConf.FullPathTest = outConf.BaseURL + outConf.ServerAdress + "/" + outConf.NewURLPref
	return outConf
}

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

	bodyOut.Result = ReductionURL(bodyIn.URL)

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

	shotURL := ReductionURL(longURL)
	w.WriteHeader(201)
	w.Write([]byte(shotURL))
}

func HandlerIndex(w http.ResponseWriter, r *http.Request) {

	idPath := r.URL.Path
	if idPath == "" {
		http.Error(w, "The parameter is missing", http.StatusBadRequest)
		return
	}

	longURL, exp := RestoreURL(idPath)

	if !exp {
		http.Error(w, "URL for the specified id was not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

}
