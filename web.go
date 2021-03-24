package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type ApiUserStats struct {
	Username string
	Minutes uint32
}

type ApiStats struct {
	Ranking []ApiUserStats 
}

var apiStats ApiStats

func addWebHandlers() {
	http.HandleFunc("/api/stats", statsHandler)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("Could not initialize webserver: ", err)
		}
	} ()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	bytes, err := json.Marshal(apiStats)
	if err != nil { log.Fatal(err) }
	w.Write(bytes)
}