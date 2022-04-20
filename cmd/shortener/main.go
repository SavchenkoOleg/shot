package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/SavchenkoOleg/shot/internal/handlers"
	"github.com/SavchenkoOleg/shot/internal/storage"
	"github.com/zenazn/goji/web"
)

type flagConfigStruct struct {
	serverAdress    string
	baseURL         string
	fileStoragePath string
}

var flagConfig flagConfigStruct

type appHandler struct {
	appContext *storage.AppContext
	h          func(*storage.AppContext, http.ResponseWriter, *http.Request)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ah.h(ah.appContext, w, r)
}

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

	r := web.New()
	r.Use(handlers.CompressGzip)

	r.Get("/"+conf.BaseURL+"/*", appHandler{&conf, handlers.HandlerIndex})
	r.Post("/", appHandler{&conf, handlers.HandlerShot})
	r.Post("/api/shorten", appHandler{&conf, handlers.HandlerShotJSON})

	err := http.ListenAndServe(conf.ServerAdress, r)
	if err != nil {
		log.Fatal(err)

	}

}
