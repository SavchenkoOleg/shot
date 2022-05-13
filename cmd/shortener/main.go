package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/SavchenkoOleg/shot/internal/handlers"
	"github.com/SavchenkoOleg/shot/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type flagConfigStruct struct {
	serverAdress       string
	baseURL            string
	fileStoragePath    string
	connectionStringDB string
}

func hendlerSetting(flags flagConfigStruct) (outConf storage.AppContext) {

	// значения по умолчанию
	outConf.NewURLPref = "newURL"
	outConf.ServerAdress = "localhost:8080"
	outConf.BaseURL = "shot"
	outConf.FileStorage = false
	outConf.FileStoragePath = ""
	outConf.UserID = ""
	outConf.ConnectionStringDB = ""

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

	сonnectionStringDB, exp := os.LookupEnv("DATABASE_DSN")
	if exp {
		outConf.ConnectionStringDB = сonnectionStringDB
	}

	// флаги, если они есть
	if flags.baseURL != "" {
		outConf.BaseURL = flags.baseURL

	}

	if flags.serverAdress != "" {
		outConf.ServerAdress = flags.serverAdress

	}

	if flags.fileStoragePath != "" {
		outConf.FileStoragePath = flags.fileStoragePath
		outConf.FileStorage = true
	}

	if flags.connectionStringDB != "" {
		outConf.ConnectionStringDB = flags.connectionStringDB
	}

	//  синтез пути, используемый в тестах
	outConf.FullPathTest = "http://" + outConf.ServerAdress + "/" + outConf.BaseURL + "/" + outConf.NewURLPref

	return outConf
}

func main() {

	var flagConfig flagConfigStruct

	// init conf
	flag.StringVar(&flagConfig.serverAdress, "a", "", "analog of environment variable SERVER_ADDRESS")
	flag.StringVar(&flagConfig.baseURL, "b", "", "analog of environment variable BASE_URL")
	flag.StringVar(&flagConfig.fileStoragePath, "f", "", "analog of environment variable FILE_STORAGE_PATH")
	flag.StringVar(&flagConfig.connectionStringDB, "d", "", "analog of environment variable FILE_STORAGE_PATH")
	flag.Parse()

	conf := hendlerSetting(flagConfig)
	conf.Ctx = context.Background()

	if conf.FileStorage {
		err := storage.RestoreMatchs(conf)
		if err != nil {
			log.Fatal(err)
		}
	}

	if conf.ConnectionStringDB != "" {
		success, err := storage.InitDBShotner(conf.Ctx, &conf)
		if err != nil {
			log.Fatal(err)

		}
		if success {
			defer conf.PgxConnect.Close(conf.Ctx)
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handlers.CompressGzip)
	r.Use(handlers.CookieMiddleware(&conf))

	r.Get("/api/user/urls", handlers.HandlerUsershortingList(&conf))
	r.Get("/"+conf.BaseURL+"/*", handlers.HandlerIndex(&conf))
	r.Get("/ping", handlers.HandlerPingDB(&conf))
	r.Post("/api/shorten/batch", handlers.HandlerShotBach(&conf))
	r.Post("/api/shorten", handlers.HandlerShotJSON(&conf))
	r.Post("/", handlers.HandlerShot(&conf))

	err := http.ListenAndServe(conf.ServerAdress, r)
	if err != nil {
		log.Fatal(err)

	}

}
