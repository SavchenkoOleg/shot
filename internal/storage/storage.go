package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
)

type FlagConfigStruct struct {
	ServerAdress    string
	BaseURL         string
	FileStoragePath string
}

type ServConfigtruct struct {
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
var ServConfig ServConfigtruct
var FlagConfig FlagConfigStruct

func HendlerSetting() (outConf ServConfigtruct) {

	// значения по умолчанию
	outConf.NewURLPref = "newURL"
	outConf.ServerAdress = ":8080"
	outConf.BaseURL = "shot"
	outConf.FileStorage = false
	outConf.FileStoragePath = ""

	// переменные окружения
	BaseURL, exp := os.LookupEnv("BASE_URL")
	if exp {
		outConf.BaseURL = BaseURL
	}

	serverAdress, exp := os.LookupEnv("SERVER_ADDRESS")
	if exp {
		outConf.ServerAdress = serverAdress
	}

	outConf.FileStoragePath, outConf.FileStorage = os.LookupEnv("FILE_STORAGE_PATH")

	// флаги, если они есть
	if FlagConfig.BaseURL != "" {
		outConf.BaseURL = FlagConfig.BaseURL

	}

	if FlagConfig.ServerAdress != "" {
		outConf.ServerAdress = FlagConfig.ServerAdress

	}

	if FlagConfig.FileStoragePath != "" {
		outConf.FileStoragePath = FlagConfig.FileStoragePath

	}

	//  синтез пути, испольхуемый в тестах
	outConf.FullPathTest = "http://localhost" + outConf.ServerAdress + "/" + outConf.BaseURL + "/" + outConf.NewURLPref

	return outConf
}

func ReductionURL(longURL string) (shotURL string, err error) {

	idURL, exp := mapLongURL[longURL]
	if !exp {

		idURL = ServConfig.NewURLPref + strconv.Itoa(len(mapLongURL)+1)

		if ServConfig.FileStorage {
			err := addMatch(longURL, idURL)
			if err != nil {
				return "", err
			}
		}

		mapLongURL[longURL] = idURL
		mapShotURL[idURL] = longURL
	}

	shotURL = "http://localhost" + ServConfig.ServerAdress + "/" + ServConfig.BaseURL + "/" + idURL

	return shotURL, nil
}

func RestoreURL(shotURL string) (restURL string, exp bool) {

	var resultURL string
	var resultExp bool

	resultURL, resultExp = mapShotURL[shotURL]

	return resultURL, resultExp
}

func addMatch(longURL, shotURL string) (err error) {

	matc := MatchEvent{longURL, shotURL}

	file, err := os.OpenFile(ServConfig.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
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

func RestoreMatchs() (err error) {

	var match MatchEvent

	file, err := os.OpenFile(ServConfig.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
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
