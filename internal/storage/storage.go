package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
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
	DelChanel          chan DelRec
}

type DelRec struct {
	UserID string
	DelURL []string
}

type MatchEvent struct {
	LongURL string `json:"longURL"`
	ShotURL string `json:"shotURL"`
}

type UsersEvent struct {
	LongURL string `json:"original_url"`
	ShotURL string `json:"short_url"`
}

type ShortenBatchIn struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchOut struct {
	CorrelationID string `json:"correlation_id"`
	ShotURL       string `json:"short_url"`
}

type userAction struct {
	userID  string
	longURL string
	shotURL string
}

var mapLongURL = make(map[string]string)
var mapShotURL = make(map[string]string)
var arrActions []userAction

func PingDB(rCtx context.Context, conf *AppContext) (exp bool) {

	ctx, cancel := context.WithTimeout(rCtx, 1*time.Second)
	defer cancel()

	err := conf.PgxConnect.Ping(ctx)
	return (err == nil)
}

func AllUserActon(сtx context.Context, conf *AppContext) (jsonText string, err error) {

	if conf.ConnectionStringDB != "" {

		return dbAllUserActon(сtx, conf)

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

func dbAllUserActon(сtx context.Context, conf *AppContext) (jsonText string, err error) {

	var userArr []UsersEvent
	var rec UsersEvent

	rows, err := conf.PgxConnect.Query(сtx, "SELECT ShotURL, LongURL FROM UserAction WHERE UserID = $1", conf.UserID)
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

func ReductionURL(ctx context.Context, longURL string, conf *AppContext) (shotURL string, err error) {

	var act userAction

	if conf.ConnectionStringDB != "" {

		return dbReductionURL(ctx, conf, longURL)

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

func dbReductionURL(ctx context.Context, conf *AppContext, longURL string) (shotURL string, err error) {

	var id int
	var URL string

	// добавляем в БД
	rows, err := conf.PgxConnect.Query(ctx, "SELECT COUNT(*) as count FROM URLs")

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

	_, err = conf.PgxConnect.Exec(ctx,
		"INSERT INTO URLs (LongURL, ShotURL, Delmark) VALUES ($1, $2, $3)",
		longURL,
		shotURL,
		false)

	if err != nil {

		if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.UniqueViolation /* or just == "23505" */ {

			rows, errqery := conf.PgxConnect.Query(ctx, "SELECT ShotURL FROM URLs WHERE LongURL = $1", longURL)

			if errqery != nil {
				return "", err
			}
			defer rows.Close()

			for rows.Next() {
				if errqery = rows.Scan(&URL); errqery == nil {
					return URL, err
				}

			}
		}
		return URL, err
	}

	conf.PgxConnect.Exec(ctx,
		"INSERT INTO UserAction (UserID, LongURL, ShotURL, idURL) VALUES ($1, $2 , $3, $4)",
		conf.UserID,
		longURL,
		shotURL,
		idURL)

	return shotURL, nil

}

func RestoreURL(ctx context.Context, conf *AppContext, shotURL string) (restURL string, exp bool, gone bool) {

	if conf.ConnectionStringDB != "" {

		return dbRestoreURL(ctx, conf, shotURL)

	}
	var resultURL string
	var resultExp bool

	resultURL, resultExp = mapShotURL[shotURL]

	return resultURL, resultExp, false
}

func dbRestoreURL(ctx context.Context, conf *AppContext, idURL string) (restURL string, exp bool, gone bool) {

	var resultURL string
	var delURL bool

	shotURL := "http://" + conf.ServerAdress + "/" + conf.BaseURL + "/" + idURL
	rows, err := conf.PgxConnect.Query(ctx, "SELECT LongURL, Delmark FROM URLs WHERE ShotURL = $1", shotURL)

	if err != nil {
		return "", false, false
	}
	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&resultURL, &delURL); err == nil {
			return resultURL, true, delURL
		}

	}

	return "", false, false
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

func InitDBShotner(ctx context.Context, conf *AppContext) (success bool, err error) {

	db, err := pgx.Connect(ctx, conf.ConnectionStringDB)
	if err != nil {
		return false, err
	}

	conf.PgxConnect = *db

	_, err = db.Exec(ctx, "Create table if not exists URLs( LongURL TEXT UNIQUE, ShotURL TEXT, СorrelationID TEXT, DelMark boolean NOT NULL)")
	if err != nil {
		return false, err
	}

	_, err = db.Exec(ctx, "Create table if not exists UserAction( UserID TEXT, LongURL TEXT, ShotURL TEXT, idURL TEXT)")
	if err != nil {

		return false, err
	}

	return true, nil
}

func DBshortenrBatch(ctx context.Context, conf *AppContext, inData []ShortenBatchIn) (outData []ShortenBatchOut, err error) {

	tx, err := conf.PgxConnect.Begin(ctx)
	if err != nil {
		return outData, err
	}
	defer tx.Rollback(ctx)

	for _, v := range inData {

		UserID := conf.UserID
		ShotURL := "http://" + conf.ServerAdress + "/" + conf.BaseURL + "/" + v.CorrelationID
		LongURL := v.OriginalURL

		_, err = tx.Exec(ctx, "INSERT INTO URLs(LongURL, ShotURL, СorrelationID, Delmark) VALUES ($1, $2 , $3, $4)", LongURL, ShotURL, v.CorrelationID, false)
		if err != nil {
			return outData, err
		}
		_, err = tx.Exec(ctx, "INSERT INTO UserAction (UserID, LongURL, ShotURL, idURL) VALUES ($1, $2, $3, $4)", UserID, LongURL, ShotURL, v.CorrelationID)
		if err != nil {
			return outData, err
		}

		rec := ShortenBatchOut{v.CorrelationID, ShotURL}
		outData = append(outData, rec)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return []ShortenBatchOut{}, err
	}

	return outData, nil

}

func ErrorCode(err error) string {
	pgerr, ok := err.(*pgconn.PgError)
	if !ok {
		return ""
	}
	return pgerr.Code

}

func bachDelete(ctx context.Context, conf *AppContext, inStruct DelRec) (err error) {

	tx, err := conf.PgxConnect.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateText :=
		`UPDATE public.urls AS url
	SET delmark = TRUE
	FROM
  	(SELECT useraction.longurl
   		FROM public.useraction AS useraction
   	JOIN
     (SELECT $1,
    	         del_idURL
    FROM unnest($2::text[]) AS del_idURL) AS del_t ON useraction.idurl = del_t.del_idURL) AS deltab
	WHERE url.longurl = deltab.longurl
  	AND NOT url.delmark`

	_, err = tx.Exec(ctx, updateText, inStruct.UserID, inStruct.DelURL)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	return err

}

func DelWorker(ctx context.Context, conf *AppContext) {
	for rec := range conf.DelChanel {
		bachDelete(ctx, conf, rec)
	}
}
