package command

import (
	"bytes"
	"fmt"
	"log"
	"waldbot/data"

	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart/v2"
)

func matesResponse(query Query) (string, *discordgo.File) {
	const matesToShow = 9

	memberId := data.ShortUserId(query.member.User.ID)
	mates, allMatesTime := data.TimeWithMates(memberId, query.dateCondition)
	
	var text string
	if query.selfUser {
		text = query.member.Mention() + ", deine Top-Freunde sind:\n"
	} else {
		text = "Die Top-Freunde von " + data.EffectiveName(query.member) + " sind:\n"
	}
	values := []chart.Value{}
	topMatesTime := uint32(0)
	listCount := matesToShow
	if len(mates) < matesToShow {
		if len(mates) == 0 {
			var text string
			if query.selfUser {
				text = "Du hast keine Freunde " + query.member.Mention() +" :cry:"
			} else {
				text = data.EffectiveName(query.member) + " hat keine Freunde :cry:"
			}
            return text, nil
		}
		listCount = len(mates)
	}
	for i := 0; i < listCount; i++ {
		mateMember, _ := data.Dc.State.Member(query.member.GuildID, data.LongUserId(mates[i].UserId))
		mateName := data.EffectiveName(mateMember)
		values = append(values, chart.Value{
			Value: float64(mates[i].Minutes) / float64(allMatesTime),
			Label: mateName,
		})
		topMatesTime += mates[i].Minutes
		text += fmt.Sprintf("%v: %v (%v)\n", data.DigitEmote(i+1), mateName, data.FormatTime(mates[i].Minutes))
	}
	values = append(values, chart.Value{
		Value: float64(allMatesTime-topMatesTime) / float64(allMatesTime),
		Label: "[Andere]",
	})
	pie := chart.PieChart{
		ColorPalette: waldColorPalette,
		Width:        1024,
		Height:       1024,
		Values:       values,
	}
	buffer := bytes.NewBuffer([]byte{})
	err := pie.Render(chart.PNG, buffer)
	if err != nil {
		log.Println("Error while creating diagram: ", err)
        return "Diagram error", nil
	}
    return text, &discordgo.File {
        Name: "mates.png",
        ContentType: "image/png",
        Reader: bytes.NewReader(buffer.Bytes()),
    }
}
