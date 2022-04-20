package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
)

type AppContext struct {
	NewURLPref      string
	ServerAdress    string
	BaseURL         string
	FullPathTest    string
	FileStorage     bool
	FileStoragePath string
}

type MatchEvent struct {
	LongURL string `json:"longURL"`
	ShotURL string `json:"shotURL"`
}

var mapLongURL = make(map[string]string)
var mapShotURL = make(map[string]string)

func ReductionURL(longURL string, conf *AppContext) (shotURL string, err error) {

	idURL, exp := mapLongURL[longURL]
	if !exp {

		idURL = conf.NewURLPref + strconv.Itoa(len(mapLongURL)+1)

		if conf.FileStorage {
			err := addMatch(longURL, idURL, conf)
			if err != nil {
				return "", err
			}
		}

		mapLongURL[longURL] = idURL
		mapShotURL[idURL] = longURL
	}

	shotURL = "http://" + conf.ServerAdress + "/" + conf.BaseURL + "/" + idURL

	return shotURL, nil
}

func RestoreURL(shotURL string) (restURL string, exp bool) {

	var resultURL string
	var resultExp bool

	resultURL, resultExp = mapShotURL[shotURL]

	return resultURL, resultExp
}

func addMatch(longURL, shotURL string, conf *AppContext) (err error) {

	matc := MatchEvent{longURL, shotURL}

	file, err := os.OpenFile(conf.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(matc)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = file.Write(data)

	return err

}

func RestoreMatchs(conf AppContext) (err error) {

	var match MatchEvent

	file, err := os.OpenFile(conf.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		json.Unmarshal(scanner.Bytes(), &match)

		mapLongURL[match.LongURL] = match.ShotURL
		mapShotURL[match.ShotURL] = match.LongURL

	}
	return err

}
