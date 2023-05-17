package main

import (
	"time"
    "waldbot/date"
    "waldbot/data"
)

type ActivityEntry struct {
	T string 	`json:"t"`
	Y uint16	`json:"y"`
}

func activity(sectionsPerDay int16, days int16, timeFormatter func(time time.Time) string) []ActivityEntry {
	minutesPerSection := int((24*60) / sectionsPerDay)
	
	if minutesPerSection * int(sectionsPerDay) != (24*60) {
		panic("sectionsPerDay has to make even section minutes")
	}
	
	sectionCount := sectionsPerDay * days
	sections := make([][]int16, sectionCount)

	dayMinute := date.DayMinute(time.Now())
	startSection := dayMinute / int16(minutesPerSection)

	startTime := time.Date(int(date.CurrentDay.Year), time.Month(date.CurrentDay.Month), int(date.CurrentDay.Day), 
		(int(startSection) * minutesPerSection) / 60, (int(startSection) * minutesPerSection) % 60, 0, 0, time.Local).
		Add(time.Hour * time.Duration(24 * -days))


	trackSession := func(dayIndex int16, date date.Date, session *data.VoiceSession, sections *[][]int16)  {
		currentMinute := 0
		for currentMinute < int(session.Minutes) {
			currentDayMinute := session.DayMinute + int16(currentMinute) 
			index := dayIndex * sectionsPerDay + (currentDayMinute / int16(minutesPerSection)) - startSection
			if index < 0 {
				currentMinute += minutesPerSection
				continue
			} else if index >= int16(sectionCount) {
				break
			}
			found := false
			for _, user := range (*sections)[index] {
				if user == session.UserID {
					found = true
					break
				}
			}
			if !found {
				(*sections)[index] = append((*sections)[index], session.UserID)
			}

			currentMinute += minutesPerSection
		}
	}

	for currentDay := int16(0); currentDay <= days; currentDay++ {
		currentDate := date.FromTime(startTime.AddDate(0, 0, int(currentDay)))
		if day, ok := data.DayData[currentDate]; ok {
			for _, sessions := range day.Channels {
				for _, session := range sessions {
					trackSession(currentDay, currentDate, &session, &sections)
				}
			}
		}
	}

	for _, session := range data.Sessions {
		trackSession(days, date.CurrentDay, &session.Session, &sections)
	} 

	output := make([]ActivityEntry, sectionCount)

	for i, section := range sections {
		sectionTime := startTime.Add(time.Minute * time.Duration(i * minutesPerSection))
		output[i] = ActivityEntry {
			T: timeFormatter(sectionTime),
			Y: uint16(len(section)),
		}
	}
	return output
}
