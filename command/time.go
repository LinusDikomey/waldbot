package command

import (
	"fmt"
	"sort"
	"waldbot/data"
	"waldbot/date"

	"github.com/bwmarrin/discordgo"
)

func timeResponse(query Query) (string, *discordgo.File) {
	id := data.ShortUserId(query.member.User.ID)
	minutes, _ := data.CalculateUserMinutes(query.dateCondition, query.dateCondition(date.CurrentDay))
	for i, user := range minutes {
		if user.UserId == id {
			// rank and time
			rankString := "Rang #" + fmt.Sprint(i+1) + " mit " + data.FormatTime(user.Minutes)
			if query.selfUser {
				rankString = (*query.member).Mention() + ", du bist " + rankString
			} else {
				rankString = data.EffectiveName(query.member) + " ist " + rankString
			}

			type ChannelMinutes struct {
				channel int16
				minutes uint32
			}

			// top channels
			channels := make([]ChannelMinutes, 0, len(user.Channels))
			for channel, minutes := range user.Channels {
				channels = append(channels, ChannelMinutes{channel: channel, minutes: minutes})
			}
			sort.Slice(channels,
				func(n, n1 int) bool {
					return channels[n].minutes > channels[n1].minutes
				})
			for i := 0; i < 9; i++ {
				if i >= len(channels) {
					break
				}
				channel, err := data.Dc.Channel(data.LongChannelId(channels[i].channel))
				name := "[Gelöschter Kanal]"
				if err == nil {
					name = channel.Name
				}
				rankString += "\n" + data.DigitEmote(i+1) + ": " + name + ": " + data.FormatTime(channels[i].minutes)
			}
			return rankString, nil
		}
	}
	return "Keine aufgezeichnete Sprachzeit für den Nutzer " + data.EffectiveName(query.member) + " gefunden", nil
}
