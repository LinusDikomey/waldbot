package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Data struct {
	ServerdatenMessageId       int64
	ServerdatenWeeklyMessageId int64
	ServerdatenDailyMessageId  int64

	ShortUserIds map[string]int16
	NextId       int16 

	ShortChannelId map[string]int16
	NextChannelId  int16

	DynamicChannels map[string][]string
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