package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	manager := NewManager()

	r.HandleFunc("/ws/{userId}", manager.serveWS)

	log.Fatal(http.ListenAndServe(":8080", r))
}
