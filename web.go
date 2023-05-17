package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"waldbot/data"
	"waldbot/oauth"
)

type ApiUserStats struct {
	Username string
	Minutes uint32
}

type ApiStats struct {
	Ranking []ApiUserStats 
}

var apiStats ApiStats

// abstraction to handle requests more easily
func handle(endpoint string, handler func(r *http.Request) (int, []byte), method string) {
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed! Use " + method))
			return
		}

		status, bytes := handler(r)
		w.WriteHeader(status)
		w.Write(bytes)
	})
}

// abstraction to handle requests more easily
func handleAuthed(endpoint string, handler func(r *http.Request, token string) (int, []byte), method string) {
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", "*")

		if r.Method == "OPTIONS" {
			w.Header().Add("Access-Control-Allow-Methods", method)
			w.Header().Add("Access-Control-Allow-Headers", "Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		} else if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed! Use " + method))
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

		status, bytes := handler(r, sessionToken)
		w.WriteHeader(status)
		w.Write(bytes)
	})
}

func addWebHandlers() {
	fmt.Printf("Starting webserver with cert file: %v and key file: %v", config.CertFile, config.KeyFile)
	//http.HandleFunc("/api/stats", statsHandler)
	handle("/api/stats", statsHandler, "GET")
	handle("/api/dayactivity", dayActivityHandler, "GET")
	handle("/api/yearactivity", yearActivityHandler, "GET")
	handle("/api/auth", authHandler, "GET")
	handleAuthed("/api/user", userHandler, "GET")
	handleAuthed("/api/logout", logoutHandler, "GET")
	go func() {
		err := http.ListenAndServeTLS(":8090", config.CertFile, config.KeyFile, nil)
		if err != nil {
			log.Fatal("Could not initialize webserver: ", err)
		}
	} ()
}

func addHeaders(w *http.ResponseWriter) {
	(*w).Header().Add("Content-Type", "application/json")
	(*w).Header().Add("Access-Control-Allow-Origin", "*")
}

func statsHandler(r *http.Request) (int, []byte) {
	bytes, _ := json.Marshal(apiStats)
	return http.StatusOK, bytes
}

func dayActivityHandler(r *http.Request) (int, []byte) {
	activity := activity(24 * 12, 1, func(t time.Time) string {
		return fmt.Sprintf("%v-%v-%v %v:%v", t.Year(), uint8(t.Month()), t.Day(), t.Hour(), t.Minute())
	})
	bytes, _ := json.Marshal(activity)
	return http.StatusOK, bytes
}

func yearActivityHandler(r *http.Request) (int, []byte) {
	activity := activity(1, 365, func(t time.Time) string {
		return fmt.Sprintf("%v-%v-%v", t.Year(), uint8(t.Month()), t.Day())
	})
	bytes, _ := json.Marshal(activity)
	return http.StatusOK, bytes
}

func authHandler(r *http.Request) (int, []byte) {
	params := r.URL.Query()
	codes, ok := params["code"]
	if !ok || len(codes) < 1 {
		return http.StatusBadRequest, []byte("Bad request, parameter 'code' missing!")
	}
	code := codes[0]
	
	body, err := oauth.AuthCode(code)
	if err != nil {
		fmt.Println("Error while making OAuth code request:", err)
		return http.StatusUnauthorized, []byte("{\"error\": \"code_invalid\"}")
	}

	sessionToken := data.AddOAuthSession(body.AccessToken, body.RefreshToken, body.ExpiresIn)

	type Response struct { 
		Token string		`json:"token"` 
	}
	response, _ := json.Marshal(Response{ Token: sessionToken })
	return http.StatusOK, response
}

func userHandler(r *http.Request, token string) (int, []byte) {
	session := data.GetOAuthSession(token)
	if session == nil {
		return http.StatusUnauthorized, []byte("Invalid session token!")
	}
	me, err := oauth.Me(session.AccessToken)
	if err != nil {
		return http.StatusInternalServerError, []byte("Request error: " + err.Error())
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
	return http.StatusOK, resp
}

func logoutHandler(r *http.Request, token string) (int, []byte) {
	if !data.RemoveOAuthSession(token) {
		return http.StatusUnauthorized, []byte("Invalid session token!")
	}
	type Response struct {
		Success string		`json:"success"`
	}
	resp, _ := json.Marshal(Response { Success: "success"})
	return http.StatusOK, resp
}
