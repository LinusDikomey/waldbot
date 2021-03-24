package main

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
)


type Date struct {
	day   uint16
	month uint8
	year  uint16
}

type sortDates []Date

func (s sortDates) Len() int {
    return len(s)
}
func (s sortDates) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s sortDates) Less(i, j int) bool {
    if s[i].year == s[j].year {
		if s[i].month == s[j].month {
			return s[i].day < s[j].day
		} else {
			return s[i].month < s[j].month
		}
	} else {
		return s[i].year < s[j].year
	}
}

func dateIsSmaller(a, b Date) bool {
    if a.year == b.year {
		if a.month == b.month {
			return a.day < b.day
		} else {
			return a.month < b.month
		}
	} else {
		return a.year < b.year
	}
}

type ChannelId = int16
type UserId = int16

type Day struct {
	channels map[ChannelId][]VoiceSession
}

type VoiceSession struct {
	dayMinute int16
	userID    UserId
	minutes   int16
}

func loadDay(buffer []byte) Day {
	reader := bytes.NewReader(buffer)
	day := Day {channels: map[ChannelId][]VoiceSession{}}
	for reader.Len() > 0 { // while bytes are still remaining
		channelId := ChannelId(0)
		len := int16(0)
		binary.Read(reader, binary.BigEndian, &channelId)
		binary.Read(reader, binary.BigEndian, &len)
		
		sessions := []VoiceSession {}
		for i := int16(0); i < len; i++ {
			session := VoiceSession {}
			binary.Read(reader, binary.BigEndian, &session.dayMinute)
			binary.Read(reader, binary.BigEndian, &session.userID)
			binary.Read(reader, binary.BigEndian, &session.minutes)
			sessions = append(sessions, session)
		}
		day.channels[channelId] = sessions
	}
	return day
}

func (day *Day) save(f *os.File) {
	for channelId, sessions := range(day.channels) {
		binary.Write(f, binary.BigEndian, &channelId)
		binary.Write(f, binary.BigEndian, int16(len(sessions)))
		
		for _, session := range(sessions) {
			binary.Write(f, binary.BigEndian, &session.dayMinute)
			binary.Write(f, binary.BigEndian, &session.userID)
			binary.Write(f, binary.BigEndian, &session.minutes)
		}
	}
}

func newDate(day uint16, month uint8, year uint16) Date {
	return Date{day: day, month: month, year: year}
}

func dateFromTime(t time.Time) Date {
	return newDate(uint16(t.Day()), uint8(t.Month()), uint16(t.Year()))
}

func readDays(dir string) {
	files, err := ioutil.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
    }
	fmt.Println("Reading", len(files), "day files!")
    for _, f := range files {
        readDay(dir, f.Name())
    }
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
	date := newDate(uint16(day), uint8(month), uint16(year))
	bytes, err := ioutil.ReadFile(path + file)
	if err != nil {
		log.Fatal("Could not read the day file: ", file, ", error: ", err)
	}
	dayData[date] = loadDay(bytes)
}