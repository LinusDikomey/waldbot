package data

import (
	"fmt"
	"time"
	"waldbot/date"
)

func MinuteUpdate(guildID string, saveFiles func()) {
	currentTime := time.Now()
	now := date.FromTime(currentTime)
	
	if _, ok := DayData[now]; !ok {
		fmt.Println("New day started!")
		DayData[now] = Day { Channels: map[int16][]VoiceSession{} }
		dayUpdate(date.CurrentDay, now)
		date.CurrentDay = now
	}
	dayMinute := date.DayMinute(currentTime)
	
	if dayMinute % 30 == 0 {
		saveFiles()
	}

	members, err := Dc.GuildMembers(guildID, "", 1000)
	if err != nil {
		panic("Could not find Guild by id specified in config: " + fmt.Sprint(err))
	}
	for _, member := range members {
		if member.User.Bot { continue }
		state, _ := Dc.State.VoiceState(guildID, member.User.ID)
		shortID := ShortUserId(member.User.ID)
		if state != nil && !state.Mute && !state.SelfMute && !state.Deaf && !state.SelfDeaf && !state.Suppress {
			// user is active in a voice chat
			
			currentChannelId := ShortChannelId(state.ChannelID)
			if active, ok := Sessions[shortID]; ok {
				if active.ChannelID == currentChannelId {
					// continue existing session
					active.Session.Minutes++
					Sessions[shortID] = active
				} else {
					//channel changed, finish existing and start new session
					DayData[now].Channels[active.ChannelID] = append(DayData[now].Channels[active.ChannelID], active.Session)

					newChannelId := currentChannelId
					Sessions[shortID] = ActiveSession {
                        Session: VoiceSession { DayMinute: dayMinute, UserID: shortID, Minutes: 1 },
                        ChannelID: newChannelId,
                    }
				}
			} else {
				//new session
				shortChannelID := currentChannelId
				Sessions[shortID] = ActiveSession{
                    Session: VoiceSession { DayMinute: dayMinute, UserID: shortID, Minutes: 1 },
                    ChannelID: shortChannelID,
                }
			}
		} else {
			// user is not in a voice chat
			if finished, ok := Sessions[shortID]; ok {
				// session just ended, append to dayData
				DayData[now].Channels[finished.ChannelID] = append(DayData[now].Channels[finished.ChannelID], finished.Session)
				delete(Sessions, shortID) 
			}
		}
	}
}

func dayUpdate(previousDay date.Date, newDay date.Date) {
	// finish sessions from previous day
	FinishSessions()
}

func FinishSessions() {
	for _, active := range(Sessions) {
		DayData[date.CurrentDay].Channels[active.ChannelID] = append(DayData[date.CurrentDay].Channels[active.ChannelID], active.Session)
	}
	Sessions = make(map[int16]ActiveSession)
}
