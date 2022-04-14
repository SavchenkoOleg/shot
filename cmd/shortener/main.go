package main

import (
	"log"
	"net/http"

	"github.com/SavchenkoOleg/shot/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	var HendlerSetting handlers.ServConfig
	HendlerSetting = handlers.HendlerSetting()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{id}", handlers.HandlerIndex)
	r.Post("/", handlers.HandlerShot)
	r.Post("/api/shorten", handlers.HandlerShotJSON)

	err := http.ListenAndServe(HendlerSetting.ServerAdress, r)
	if err != nil {
		log.Fatal(err)

	}

}
