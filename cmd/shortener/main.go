package main

import (
	"log"
	"net/http"

	"github.com/SavchenkoOleg/shot.git/internal/handlers"
)

func firstResort(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		handlers.HandlerShot(w, r)
	case http.MethodGet:
		handlers.HandlerIndex(w, r)
	default:
		http.Error(w, "Only POST or GET requests are allowed!", http.StatusMethodNotAllowed)
	}

}
func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", firstResort)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}

}
