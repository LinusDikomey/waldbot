package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
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
	http.HandleFunc("/api/dayactivity", dayActivityHandler)
	http.HandleFunc("/api/yearactivity", yearActivityHandler)
	go func() {
		err := http.ListenAndServeTLS(":8080", config.CertFile, config.KeyFile, nil)
		if err != nil {
			log.Fatal("Could not initialize webserver: ", err)
		}
	} ()
}

func addHeaders(w *http.ResponseWriter) {
	(*w).Header().Add("Content-Type", "application/json")
	(*w).Header().Add("Access-Control-Allow-Origin", "*")
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(&w)
	bytes, err := json.Marshal(apiStats)
	if err != nil { log.Fatal(err) }
	w.Write(bytes)
}

func dayActivityHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(&w)
	
	activity := activity(24 * 12, 1, func(t time.Time) string {
		return fmt.Sprintf("%v-%v-%v %v:%v", t.Year(), uint8(t.Month()), t.Day(), t.Hour(), t.Minute())
	})
	bytes, err := json.Marshal(activity)
	if err != nil { log.Fatal(err) }
	w.Write(bytes)
}

func yearActivityHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(&w)
	
	activity := activity(1, 365, func(t time.Time) string {
		return fmt.Sprintf("%v-%v-%v", t.Year(), uint8(t.Month()), t.Day())
	})
	bytes, err := json.Marshal(activity)
	if err != nil { log.Fatal(err) }
	w.Write(bytes)
}