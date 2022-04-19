package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/SavchenkoOleg/shot/internal/conf"
	"github.com/SavchenkoOleg/shot/internal/handlers"
	"github.com/SavchenkoOleg/shot/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func init() {

	flag.StringVar(&conf.FlagConfig.ServerAdress, "a", "", "analog of environment variable SERVER_ADDRESS")
	flag.StringVar(&conf.FlagConfig.BaseURL, "b", "", "analog of environment variable BASE_URL")
	flag.StringVar(&conf.FlagConfig.FileStoragePath, "f", "", "analog of environment variable FILE_STORAGE_PATH")
}

func main() {

	flag.Parse()
	conf.ServConfig = conf.HendlerSetting()

	if conf.ServConfig.FileStorage {
		err := storage.RestoreMatchs()
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

	r.Get("/"+conf.ServConfig.BaseURL+"/{id}", handlers.HandlerIndex)
	r.Post("/", handlers.HandlerShot)
	r.Post("/api/shorten", handlers.HandlerShotJSON)

	err := http.ListenAndServe(conf.ServConfig.ServerAdress, r)
	if err != nil {
		log.Fatal(err)

	}

}
