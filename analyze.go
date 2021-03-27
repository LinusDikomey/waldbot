package main

import (
	"time"
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

	dayMinute := dayMinute(time.Now())
	startSection := dayMinute / int16(minutesPerSection)

	startTime := time.Date(int(currentDay.year), time.Month(currentDay.month), int(currentDay.day), 
		(int(startSection) * minutesPerSection) / 60, (int(startSection) * minutesPerSection) % 60, 0, 0, time.Local).
		Add(time.Hour * time.Duration(24 * -days))


	trackSession := func(dayIndex int16, date Date, session *VoiceSession, sections *[][]int16)  {
		currentMinute := 0
		for currentMinute < int(session.minutes) {
			currentDayMinute := session.dayMinute + int16(currentMinute) 
			index := dayIndex * sectionsPerDay + (currentDayMinute / int16(minutesPerSection)) - startSection
			if index < 0 {
				currentMinute += minutesPerSection
				continue
			} else if index >= int16(sectionCount) {
				break
			}
			found := false
			for _, user := range (*sections)[index] {
				if user == session.userID {
					found = true
					break
				}
			}
			if !found {
				(*sections)[index] = append((*sections)[index], session.userID)
			}

			currentMinute += minutesPerSection
		}
	}

	for currentDay := int16(0); currentDay <= days; currentDay++ {
		currentDate := dateFromTime(startTime.AddDate(0, 0, int(currentDay)))
		if day, ok := dayData[currentDate]; ok {
			for _, sessions := range day.channels {
				for _, session := range sessions {
					trackSession(currentDay, currentDate, &session, &sections)
				}
			}
		}
	}

	for _, session := range sessions {
		trackSession(days, currentDay, &session.session, &sections)
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