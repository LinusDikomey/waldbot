package command

import (
	"fmt"
	"waldbot/data"

	"github.com/bwmarrin/discordgo"
)

func sessionResponse(query Query) (string, *discordgo.File) {
	if session, ok := data.Sessions[data.ShortUserId(query.member.User.ID)]; ok {
		voiceChannel, _ := data.Dc.State.Channel(data.LongChannelId(session.ChannelID))

		return fmt.Sprintf("Du bist seit %v Minuten im Kanal %v", session.Session.Minutes, voiceChannel.Name), nil
	} else {
		return "Du bist momentan in keinem Sprachkanal", nil
	}
}
