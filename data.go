package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/LinusDikomey/waldbot/oauth"
)

type OAuthLogin struct {
	AccessToken string
	RefreshToken string
	AccessExpires uint64
	Expires uint64
}

type Data struct {
	ServerdatenMessageId       int64
	ServerdatenWeeklyMessageId int64
	ServerdatenDailyMessageId  int64

	ShortUserIds map[string]int16
	NextId       int16 

	ShortChannelId map[string]int16
	NextChannelId  int16

	DynamicChannels map[string][]string

	OAuthLogins map[string]OAuthLogin
}

var (
	data Data
)

func loadData() {
	fmt.Println("Loading data file...")
	file, err := os.Open("./data/data.json")
	if err != nil {
		log.Fatal("could not find data.json: ", err)
	}
	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)
	err = json.Unmarshal(bytes, &data)
	if data.OAuthLogins == nil {
		data.OAuthLogins = map[string]OAuthLogin {}
	}
}

func saveData() {
	fmt.Println("Saving data file...")
	file, _ := os.Create("./data/data.json")
	bytes, err := json.Marshal(&data)
	if err != nil {
		log.Fatal("Json Marshal error:", err)
	}
	_, err = file.Write(bytes)
	if err != nil {
		log.Fatal("Could not save json: ", err)
	}
	file.Close()
}

func shortUserId(id string) int16 {
	if val, ok := data.ShortUserIds[id]; ok {
		return val
	}
	shortId := data.NextId
	data.NextId++
	data.ShortUserIds[fmt.Sprint(id)] = shortId
	return shortId
}

func shortChannelId(id string) int16 {
	if val, ok := data.ShortChannelId[id]; ok {
		return val
	}
	shortId := data.NextChannelId
	data.NextChannelId++
	data.ShortChannelId[fmt.Sprint(id)] = shortId
	return shortId
}

func longChannelId(id int16) string {
	for longId, shortId := range data.ShortChannelId {
		if shortId == id {
			return longId
		}
	}
	panic("Could not find long channel id by short id: " + fmt.Sprint(id))
}

func longUserId(id int16) string {
	for longId, shortId := range data.ShortUserIds {
		if shortId == id {
			return longId
		}
	}
	panic("Could not find long user id by short id: " + fmt.Sprint(id))
}

func addOAuthSession(access string, refresh string, expiresIn uint64) string {
	found := true
	var token string
	// search for unique token
	for found {
		token = randomBase64String(32)
		_, found = data.OAuthLogins[token]
	}
	now := time.Now()
	data.OAuthLogins[token] = OAuthLogin {
		AccessToken: access,
		RefreshToken: refresh,
		AccessExpires: uint64(now.Add(time.Duration(expiresIn) * time.Second).Unix()),
		Expires: uint64(now.Add(24 * time.Hour * 30).Unix()), // session expires after 30 days
	}
	return token
}

func getOAuthSession(token string) *OAuthLogin {
	if session, ok := data.OAuthLogins[token]; ok {
		now := time.Now()
		// update expiry
		session.Expires = uint64(now.Add(24 * time.Hour * 90).Unix())

		if session.AccessExpires < uint64(now.Unix()) + 60 {
			fmt.Println("Refreshing OAuth2 access token")
			access, err := oauth.RefreshToken(session.RefreshToken)
			if err != nil {
				fmt.Println("Error while refreshing token:", err)
				return nil
			}
			session.AccessToken = access.AccessToken
			session.AccessExpires = uint64(now.Add(time.Duration(access.ExpiresIn) * time.Second).Unix())
		}

		data.OAuthLogins[token] = session
		// return found session
		return &session
	} else {
		return nil
	}
}

func removeOAuthSession(token string) bool {
	if _, ok := data.OAuthLogins[token]; ok {
		delete(data.OAuthLogins, token)
		return true
	}
	return false
}