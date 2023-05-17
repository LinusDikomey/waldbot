package command

import (
	"bytes"
	"log"
	"waldbot/data"
	"waldbot/date"

	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart/v2"
)

func channelsResponse(query Query) (string, *discordgo.File) {
	shortUser := data.ShortUserId(query.member.User.ID)

	stats, _ := data.CalculateUserMinutes(query.dateCondition, query.dateCondition(date.CurrentDay))

	for _, userStats := range stats {
		if userStats.UserId != shortUser {
			continue
		}

		values := make([]chart.Value, len(userStats.Channels))
		for id, minutes := range userStats.Channels {
			name := "[Gelöschter Kanal]"
			// don't label channels with small pieces
			if (userStats.Minutes / minutes) < 50 {
				if channel, ok := data.Dc.Channel(data.LongChannelId(id)); ok == nil {
					name = channel.Name
				}
			} else {
				name = ""
			}

			values = append(values, chart.Value{
				Value: float64(minutes),
				Label: name,
			})
		}
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
            return "Fehler beim Erstellen des Diagramm: " + err.Error(), nil 
		}
		return "", &discordgo.File {
            Name: "channels.png",
            ContentType: "image/png",
            Reader: bytes.NewReader(buffer.Bytes()),
            
        }
	}
    return "Keine Daten gefunden!", nil
}
