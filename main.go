package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	manager := NewManager()

	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	r.HandleFunc("/ws/{userId}", manager.serveWS)

	log.Fatal(http.ListenAndServe(":8080", r))
}
