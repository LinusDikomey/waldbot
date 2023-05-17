package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"waldbot/date"
)



type ChannelId = int16
type UserId = int16

type Day struct {
	Channels map[ChannelId][]VoiceSession
}

type VoiceSession struct {
	DayMinute int16
	UserID    UserId
	Minutes   int16
}

func loadDay(buffer []byte) Day {
	reader := bytes.NewReader(buffer)
	day := Day {Channels: map[ChannelId][]VoiceSession{}}
	for reader.Len() > 0 { // while bytes are still remaining
		channelId := ChannelId(0)
		len := int16(0)
		binary.Read(reader, binary.BigEndian, &channelId)
		binary.Read(reader, binary.BigEndian, &len)
		
		sessions := []VoiceSession {}
		for i := int16(0); i < len; i++ {
			session := VoiceSession {}
			binary.Read(reader, binary.BigEndian, &session.DayMinute)
			binary.Read(reader, binary.BigEndian, &session.UserID)
			binary.Read(reader, binary.BigEndian, &session.Minutes)
			sessions = append(sessions, session)
		}
		day.Channels[channelId] = sessions
	}
	return day
}

func (day *Day) Save(f *os.File) {
	for channelId, sessions := range(day.Channels) {
		binary.Write(f, binary.BigEndian, &channelId)
		binary.Write(f, binary.BigEndian, int16(len(sessions)))
		
		for _, session := range(sessions) {
			binary.Write(f, binary.BigEndian, &session.DayMinute)
			binary.Write(f, binary.BigEndian, &session.UserID)
			binary.Write(f, binary.BigEndian, &session.Minutes)
		}
	}
}

func ReadDays(dir string) {
	files, err := ioutil.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
    }
	fmt.Println("Reading", len(files), "day files!")
	start_time := time.Now()
    for _, f := range files {
        readDay(dir, f.Name())
    }
	fmt.Println("Took", time.Since(start_time), "to load files")
}

func readDay(path string, file string) {
	split := strings.Split(file, ".");
	if split[1] != "day" {
		log.Fatal("Invalid file found in 'days' subfolder: ", file)
	}
	dateStrings := strings.Split(split[0], "-")
	day, err1 := strconv.Atoi(dateStrings[0])
	month, err2 := strconv.Atoi(dateStrings[1])
	year, err3 := strconv.Atoi(dateStrings[2])
	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal("Could not parse day of days file '", file, "'")
	}
	date := date.New(uint16(day), uint8(month), uint16(year))
	bytes, err := ioutil.ReadFile(path + file)
	if err != nil {
		log.Fatal("Could not read the day file: ", file, ", error: ", err)
	}
	DayData[date] = loadDay(bytes)
}
