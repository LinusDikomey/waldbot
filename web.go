package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/LinusDikomey/waldbot/oauth"
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
	http.HandleFunc("/api/auth", authHandler)
	http.HandleFunc("/api/user", userHandler)
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
	addHeaders(&w)

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed! Use GET!"))
		return
	}
	params := r.URL.Query()
	codes, ok := params["code"]
	if !ok || len(codes) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request, parameter 'code' missing!"))
		return
	}
	code := codes[0]
	
	body, err := oauth.AuthCode(code)
	if err != nil {
		fmt.Println("Error while making OAuth code request:", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\"error\": \"code_invalid\"}"))
		return
	}

	sessionToken := addOAuthSession(body.AccessToken, body.RefreshToken, body.ExpiresIn)

	type Response struct { 
		Token string		`json:"token"` 
	}
	response, _ := json.Marshal(Response{ Token: sessionToken })
	w.Write(response)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(&w)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed! Use GET!"))
		return
	}
	
	auth := r.Header["Authorization"]
	const AUTH_PREFIX = "Bearer "
	if len(auth) < 1 || !strings.HasPrefix(auth[0], AUTH_PREFIX) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized! Use Bearer token"))
		return
	}
	sessionToken := auth[0][len(AUTH_PREFIX):]
	session := getOAuthSession(sessionToken)
	if session == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Invalid session token!"))
		return
	}
	me, err := oauth.Me(session.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Request error: " + err.Error()))
		return
	}

	type Response struct {
		Username string			`json:"username"`
		Discriminator string	`json:"discriminator"`
		Id string				`json:"id"`
		AvatarURL string		`json:"avatarUrl"`
	}

	resp, _ := json.Marshal(Response {
		Username: me.User.Username,
		Discriminator: me.User.Discriminator,
		Id: me.User.ID,
		AvatarURL: me.User.AvatarURL(""),
	})
	w.Write(resp)
}