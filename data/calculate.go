package data

import (
    "sort"

    "waldbot/date"
)

// returns sorted list of UserStats
func CalculateUserMinutes(dayCondition func(date.Date) bool, includeActive bool) ([]UserStats, int) {
	dayCount := 0
	users := map[int16]UserStats {}
	addSession := func(channel int16, session VoiceSession) {
		// get or create stats for user
		var userStats UserStats
		if stats, ok := users[session.UserID]; ok {
			userStats = stats
		} else {
			userStats = UserStats{UserId: session.UserID, Minutes: 0, Channels: map[ChannelId]uint32 {}}
		}
		// increase overall minutes
		userStats.Minutes += uint32(session.Minutes)
		// get or initialize stats for channel
		channelMinutes := uint32(0)
		if minutes, ok := userStats.Channels[channel]; ok {
			channelMinutes = minutes
		}
		// increase minutes for channel
		userStats.Channels[channel] = channelMinutes + uint32(session.Minutes)
		// put userStats back into users map
		users[session.UserID] = userStats
	}

	for date, day := range DayData {
		if dayCondition(date) {
			dayCount++
			for channel, sessions := range day.Channels {
				for _, session := range sessions {
					addSession(channel, session)
				}
			}
		}
	}
	if includeActive {
		for _, active := range Sessions {
			addSession(active.ChannelID, active.Session)
		}
	}
	// put values into list and sort
	list := make([]UserStats, 0, len(users))
	for _, val := range users {
		list = append(list, val)
	}
	f := func (n, n1 int) bool {
		return list[n].Minutes > list[n1].Minutes
	}
	sort.Slice(list, f)
	return list, dayCount
}

type UserStats struct {
	UserId int16
	Minutes uint32
	Channels map[ChannelId]uint32 
}

type SortUserStats []UserStats

func (s SortUserStats) Len() int {
	return len(s)
}
func (s SortUserStats) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortUserStats) Less(i, j int) bool {
	return s[i].Minutes < s[j].Minutes
}

func TimeWithMates(user int16, dateCondition date.DateCondition) ([]UserStats, uint32) {
	min := func(a, b int32) int32 {
		if a < b { return a } else { return b }
	}
	mates := map[int16]UserStats {}
	allTime := uint32(0)
	for day, data := range DayData {
		if !dateCondition(day) { continue }
		for _, sessions := range data.Channels {
			for _, session := range sessions {
				if session.UserID != user { continue }
				for _, otherSession := range sessions {
					if otherSession.UserID == user { continue }
					var minutes int32
					a := int32(session.DayMinute)
					b := a + int32(session.Minutes)
					c := int32(otherSession.DayMinute)
					d := c + int32(otherSession.Minutes)
					if a < c {
						minutes = min(b-c, d-c)
					} else {
						minutes = min(d-a, b-a)
					}
					if minutes > 0 {
						var stats UserStats
						if found, ok := mates[otherSession.UserID]; ok {
							stats = found
						} else {
							stats = UserStats {UserId: otherSession.UserID, Minutes: 0} 
						}
						stats.Minutes += uint32(minutes)
						mates[otherSession.UserID] = stats

						allTime += uint32(minutes)
					}
				}
			}
		}
	}
	slice := make([]UserStats, len(mates))
	index := 0
	for _, entry := range mates {
		slice[index] = entry
		index++
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Minutes > slice[j].Minutes
	})
	return slice, allTime
}
