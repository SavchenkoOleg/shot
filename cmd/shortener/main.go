package main

import (
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
	serverAdress    string
	baseURL         string
	fileStoragePath string
}

var flagConfig flagConfigStruct

func hendlerSetting(flags flagConfigStruct) (outConf storage.AppContext) {

	// значения по умолчанию
	outConf.NewURLPref = "newURL"
	outConf.ServerAdress = "localhost:8080"
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

	//  синтез пути, используемый в тестах
	outConf.FullPathTest = "http://" + outConf.ServerAdress + "/" + outConf.BaseURL + "/" + outConf.NewURLPref

	return outConf
}

func main() {

	// init conf
	flag.StringVar(&flagConfig.serverAdress, "a", "", "analog of environment variable SERVER_ADDRESS")
	flag.StringVar(&flagConfig.baseURL, "b", "", "analog of environment variable BASE_URL")
	flag.StringVar(&flagConfig.fileStoragePath, "f", "", "analog of environment variable FILE_STORAGE_PATH")
	flag.Parse()

	conf := hendlerSetting(flagConfig)

	if conf.FileStorage {
		err := storage.RestoreMatchs(conf)
		if err != nil {
			log.Fatal(err)
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handlers.CompressGzip)

	r.Get("/"+conf.BaseURL+"/*", handlers.HandlerIndex(&conf))
	r.Post("/", handlers.HandlerShot(&conf))
	r.Post("/api/shorten", handlers.HandlerShotJSON(&conf))

	err := http.ListenAndServe(conf.ServerAdress, r)
	if err != nil {
		log.Fatal(err)

	}

}
