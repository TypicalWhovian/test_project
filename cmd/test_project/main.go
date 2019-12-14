package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"test_project/internal"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", internal.Handler).Methods(http.MethodPost)
	http.Handle("/", r)
	log.Print("listening on 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
