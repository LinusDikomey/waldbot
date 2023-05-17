package command

import (
	"bytes"
	"log"
	"sort"
	"time"
	"waldbot/data"
	"waldbot/date"

	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart/v2"
)

func usercountResponse(query Query) (string, *discordgo.File) {
	dates := make([]date.Date, 0, len(data.DayData))

	for date := range data.DayData {
		dates = append(dates, date)
	}

	sort.Slice(dates, func(i, j int) bool {
		return date.IsSmaller(dates[i], dates[j])
	})

	var xValues []time.Time
	var yValues []float64
	for _, date := range dates {
		if !query.dateCondition(date) {
			continue
		}
		var users []int16
		for _, channel := range data.DayData[date].Channels {
			for _, session := range channel {
				for _, user := range users {
					if session.UserID != user {
						users = append(users, session.UserID)
					}
				}
			}
		}
		xValues = append(xValues, time.Date(int(date.Year), time.Month(date.Month), int(date.Day), 0, 0, 0, 0, time.Local))
		yValues = append(yValues, float64(len(users)))
	}
	if len(xValues) < 2 {
		return "Nicht genug Daten gefunden!", nil
	}

	graph := chart.Chart{
		Width:  1280,
		Height: 720,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    80,
				Left:   10,
				Right:  10,
				Bottom: 10,
			},
		},
		ColorPalette: waldColorPalette,
		Title:        "User über Zeit",
		XAxis: chart.XAxis{
			Name:           "Datum",
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:           "User",
			ValueFormatter: chart.IntValueFormatter,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: xValues,
				YValues: yValues,
			},
		},
	}
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		log.Println("Error while creating diagram: ", err)
        return "Diagram error", nil
	}
    return "", &discordgo.File {
        Name: "usercount.png",
        ContentType: "image/png",
        Reader: bytes.NewReader(buffer.Bytes()),
    }
}
