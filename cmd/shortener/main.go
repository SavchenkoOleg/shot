package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/SavchenkoOleg/shot/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	HendlerSetting := handlers.HendlerSetting()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{id}", handlers.HandlerIndex)
	r.Post("/", handlers.HandlerShot)
	r.Post("/api/shorten", handlers.HandlerShotJSON)

	addr := strings.Replace(HendlerSetting.ServerAdress, "http://", "", 1)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal(err)

	}

}
