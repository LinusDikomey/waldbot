package command

import (
	"math"
	"waldbot/data"
)

func hoursResponse(query Query) string {
	min := func(a, b int16) int16 {
		if a < b {
			return a
		} else {
			return b
		}
	}

	const SECTION_SIZE = 5
	const SECTIONS = (24 * 60) / SECTION_SIZE

	shortUser := data.ShortUserId(query.member.User.ID)
	minutesSum := uint32(0)
	minutesBySections := make([]uint32, SECTIONS)

	dayCount := 0
	for date, day := range data.DayData {
		if !query.dateCondition(date) {
			continue
		}
		dayCount++
		for _, sessions := range day.Channels {
			for _, session := range sessions {
				if session.UserID != shortUser {
					continue
				}
				// session from requested user
				currentMinute := session.DayMinute
				remaining := session.Minutes
				for remaining > 0 {
					section := currentMinute / SECTION_SIZE
					if section >= SECTIONS {
						section = SECTIONS - 1
					}
					minutesLeftInSection := min(remaining, SECTION_SIZE-(currentMinute%SECTION_SIZE))
					minutesBySections[section] += uint32(minutesLeftInSection)
					minutesSum += uint32(minutesLeftInSection)
					currentMinute += minutesLeftInSection
					remaining -= minutesLeftInSection
				}
			}
		}
	}

	if minutesSum == 0 {
		return "Keine Daten gefunden!"
	}

	xAxis := make([]float64, SECTIONS)
	yAxis := make([]float64, SECTIONS)

	maxMinutesPerSection := float64(SECTION_SIZE * dayCount)

	maxY := float64(0)

	for i := 0; i < SECTIONS; i++ {
		xAxis[i] = float64(i * SECTION_SIZE)
		value := float64(minutesBySections[i]) / maxMinutesPerSection
		maxY = math.Max(maxY, value)
		yAxis[i] = value
	}

	dc.ChannelFileSend(channel, "diagram.png", bytes.NewReader(
		dayTimeChart(
			fmt.Sprintf("Sprachchat-Zeitverteilung von '%v'", effectiveName(member)),
			xAxis,
			yAxis,
			maxY,
		)))
}
