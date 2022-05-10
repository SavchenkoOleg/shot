package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

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

func PingDB(conf *AppContext) (exp bool) {

	db, err := pgx.Connect(context.Background(), conf.ConnectionStringDB)
	if err != nil {

		return false
	}

	defer db.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	return (err == nil)
}

func AllUserActon(conf *AppContext) (jsonText string, err error) {

	if conf.ConnectionStringDB != "" {

		return dbAllUserActon(conf)

	}

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

func dbAllUserActon(conf *AppContext) (jsonText string, err error) {

	var userArr []UsersEvent
	var rec UsersEvent
	ctx := context.Background()

	rows, err := conf.PgxConnect.Query(ctx, "SELECT ShotURL, LongURL FROM UserAction WHERE UserID = $1", conf.UserID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&rec.ShotURL, &rec.LongURL)
		if err == nil {
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

	if conf.ConnectionStringDB != "" {

		return dbReductionURL(longURL, conf)

	}

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

func dbReductionURL(longURL string, conf *AppContext) (shotURL string, err error) {

	var id int
	var URL string
	ctx := context.Background()

	rows, err := conf.PgxConnect.Query(ctx, "SELECT ShotURL FROM URLs WHERE LongURL = $1", longURL)

	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&URL); err == nil {
			return URL, err
		}

	}

	// нет записи соответсвия "longURL"
	// добавляем в БД
	rows, err = conf.PgxConnect.Query(ctx, "SELECT COUNT(*) as count FROM URLs")

	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return "", err
		}
	}

	idURL := conf.NewURLPref + strconv.Itoa(id+1)
	shotURL = "http://" + conf.ServerAdress + "/" + conf.BaseURL + "/" + idURL

	conf.PgxConnect.Exec(ctx,
		"INSERT INTO URLs (LongURL, ShotURL) VALUES ($1, $2)",
		longURL,
		shotURL)
	conf.PgxConnect.Exec(ctx,
		"INSERT INTO UserAction (UserID, LongURL, ShotURL) VALUES ($1, $2 , $3)",
		conf.UserID,
		longURL,
		shotURL)

	return shotURL, nil

}

func RestoreURL(shotURL string, conf *AppContext) (restURL string, exp bool) {

	if conf.ConnectionStringDB != "" {

		return dbRestoreURL(shotURL, conf)

	}
	var resultURL string
	var resultExp bool

	resultURL, resultExp = mapShotURL[shotURL]

	return resultURL, resultExp
}

func dbRestoreURL(idURL string, conf *AppContext) (restURL string, exp bool) {

	var resultURL string
	ctx := context.Background()

	shotURL := "http://" + conf.ServerAdress + "/" + conf.BaseURL + "/" + idURL
	rows, err := conf.PgxConnect.Query(ctx, "SELECT LongURL FROM URLs WHERE ShotURL = $1", shotURL)

	if err != nil {
		return "", false
	}
	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&resultURL); err == nil {
			return resultURL, true
		}

	}

	return "", false
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

func InitDBShotner(conf *AppContext) (success bool, err error) {

	db, err := pgx.Connect(context.Background(), conf.ConnectionStringDB)
	if err != nil {
		return false, err
	}

	conf.PgxConnect = *db
	ctx := context.Background()

	_, err = db.Exec(ctx, "Create table if not exists URLs( LongURL TEXT, ShotURL TEXT)")
	if err != nil {
		return false, err
	}

	_, err = db.Exec(ctx, "Create table if not exists UserAction( UserID TEXT, LongURL TEXT, ShotURL TEXT)")
	if err != nil {

		return false, err
	}

	return true, nil
}
