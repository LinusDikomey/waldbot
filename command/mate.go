package command

import (
	"fmt"
	"waldbot/data"

	"github.com/bwmarrin/discordgo"
)

func mateResponse(query Query) (string, *discordgo.File) {
    if query.member == query.mate {
        return "Lol du willst den bot aber auch kaputt machen, oder?", nil
    }

	mateId := data.ShortUserId(query.mate.User.ID)
	authorId := data.ShortUserId(query.member.User.ID)
	mates, _ := data.TimeWithMates(authorId, query.dateCondition)
	for i, mate := range mates {
		if mate.UserId == mateId {
            if query.selfUser {
                return fmt.Sprintf(
                    "%v, deine überlappende Zeit mit %v beträgt %v (Platz %v)",
                    query.member.Mention(), data.EffectiveName(query.mate), data.FormatTime(mate.Minutes), i+1,
                ), nil
            } else {
                a := data.EffectiveName(query.member)
                b := data.EffectiveName(query.mate)
                return fmt.Sprintf(
                    "Die überlappende Zeit von %v mit %v beträgt %v (Platz %v von %v)",
                    a, b, data.FormatTime(mate.Minutes), i+1, a,
                ), nil
            }
		}
	}
    return "Keine überlappende Zeit gefunden", nil
}
