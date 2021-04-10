package main

import (
	"bytes"
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
	http.HandleFunc("api/auth", authHandler)
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

func authHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ClientId string			`json:"client_id"`
		ClientSecret string		`json:"client_secret"`
		Code string				`json:"code"`
		RedirectUri string		`json:"redirect_uri"`
		Scope string			`json:"scope"`
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed! Use GET!"))
		return
	}
	params := r.URL.Query()
	codes, ok := params["code"]
	code := codes[0]
	if !ok || len(code) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request, parameter 'code' missing!"))
	}
	fmt.Println("Received OAuth2 code:", code)
	
	data, _ := json.Marshal(Request {
		ClientId: oauthClientId,
		ClientSecret: oauthClientSecret,
		Code: code,
		RedirectUri: "wald.mbehrmann.de/auth",
		Scope: "identify",
	})
	req, err := http.NewRequest("POST", "https://discord.com/api/v8/oauth2/token", bytes.NewReader(data))
	if err != nil {
		fmt.Println("could not create request:", err)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client {}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while executing auth request:", err)
		return
	}
	fmt.Println("Response:", resp)
	
}