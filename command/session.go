package command

import (
	"fmt"
	"waldbot/data"

	"github.com/bwmarrin/discordgo"
)

func sessionResponse(query Query) (string, *discordgo.File) {
    var address string
    if query.selfUser {
        address = "Du bist"
    } else {
        address = data.EffectiveName(query.member) + " ist"
    }
	if session, ok := data.Sessions[data.ShortUserId(query.member.User.ID)]; ok {
		voiceChannel, _ := data.Dc.State.Channel(data.LongChannelId(session.ChannelID))

		return fmt.Sprintf("%v seit %v Minuten im Kanal %v", address, session.Session.Minutes, voiceChannel.Name), nil
	} else {
		return address + " momentan in keinem Sprachkanal", nil
	}
}
