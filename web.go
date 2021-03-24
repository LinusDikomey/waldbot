package main

import (
	"encoding/json"
	"fmt"
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
	fmt.Printf("Starting webserver with cert file: %v and key file: %v", config.CertFile, config.KeyFile)
	http.HandleFunc("/api/stats", statsHandler)
	go func() {
		err := http.ListenAndServeTLS(":8080", config.CertFile, config.KeyFile, nil)
		if err != nil {
			log.Fatal("Could not initialize webserver: ", err)
		}
	} ()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	bytes, err := json.Marshal(apiStats)
	if err != nil { log.Fatal(err) }
	w.Write(bytes)
}