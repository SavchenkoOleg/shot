package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4"
)

type AppContext struct {
	NewURLPref         string
	ServerAdress       string
	BaseURL            string
	FullPathTest       string
	FileStorage        bool
	FileStoragePath    string
	UserID             string
	ConnectionStringDB string
	PgxConnect         pgx.Conn
}

type MatchEvent struct {
	LongURL string `json:"longURL"`
	ShotURL string `json:"shotURL"`
}

type UsersEvent struct {
	LongURL string `json:"original_url"`
	ShotURL string `json:"short_url"`
}

type userAction struct {
	userID  string
	longURL string
	shotURL string
}

var mapLongURL = make(map[string]string)
var mapShotURL = make(map[string]string)
var arrActions []userAction

func AllUserActon(conf *AppContext) (jsonText string, err error) {

	var userArr []UsersEvent

	for i := 0; i < len(arrActions); i++ {
		if arrActions[i].userID == conf.UserID {
			rec := UsersEvent{arrActions[i].longURL, arrActions[i].shotURL}
			userArr = append(userArr, rec)
		}
	}

	if len(userArr) == 0 {
		return "", nil
	}

	data, err := json.MarshalIndent(userArr, "", "")
	if err != nil {
		return "", err
	}

	return string(data), nil

}

func ReductionURL(longURL string, conf *AppContext) (shotURL string, err error) {

	var act userAction

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

	act.userID = conf.UserID
	act.longURL = longURL
	act.shotURL = shotURL

	arrActions = append(arrActions, act)

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
