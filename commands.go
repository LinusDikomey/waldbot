package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart/v2"
)

type Command struct {
	prefix      string
	handler     func(args string, channel string, author *discordgo.Member)
	description string
}

var commands = [...]Command{
	{prefix: "ping", handler: pingCommand},
	{prefix: "help", handler: helpCommand},
	{prefix: "time", handler: timeCommand},
	{prefix: "hours", handler: hoursCommand},
	{prefix: "channels", handler: channelsCommand},
	{prefix: "session", handler: sessionHandler},
	{prefix: "mate", handler: mateHandler},
	{prefix: "mates", handler: matesHandler},
	{prefix: "stonks", handler: stonksHandler},
}

// this is stupid but it looks like it has to be done to avoid initialization cycle errors
var commandDescriptions = [...]Command{
	{prefix: "ping", description: "Pong"},
	{prefix: "help", description: "Zeigt diese Hilfe an"},
	{prefix: "time {optional: Nutzer/Zeitfenster}", description: "Zeigt Sprachchatzeit, Rang und Lieblingskanäle an"},
	{prefix: "hours {optional: Nutzer/Zeitfenster}", description: "Generiert ein Diagram mit der Zeitverteilung über einen Zeitraum"},
	{prefix: "channels {optional: Nutzer/Zeitfenster}", description: "Zeigt ein Tortendiagramm mit der Verteilung deiner genutzten Kanäle"},
	{prefix: "session", description: "Zeigt deine aktuelle Voicechat-Session"},
	{prefix: "mate {Nutzer}", description: "Zeigt deine Zeit mit einem Nutzer an"},
	{prefix: "mates", description: "Zeigt ein Tortendiagramm der Sprachchatzeit mit anderen Nutzern"},
	{prefix: "stonks {optional: Nutzer/Zeitfenster}", description: "GME TO THE MOON"},
}

func pingCommand(content string, channel string, author *discordgo.Member) {
	dc.ChannelMessageSend(channel, "pong!")
}

func helpCommand(content string, channel string, author *discordgo.Member) {
	text := "**Liste aller Befehle:**\n"
	for _, command := range commandDescriptions {
		text += "**!" + command.prefix + "**: " + command.description + "\n"
	}
	dc.ChannelMessageSend(channel, text)
}

func timeCommand(args string, channel string, author *discordgo.Member) {
	//authorId := shortUserId(author.User.ID)

	ok, dateCondition, member := parseUserOrTime(args, channel, author)
	if !ok {
		return
	} // error messages handled by util function, just return

	id := shortUserId(member.User.ID)
	minutes, _ := calculateUserMinutes(dateCondition, dateCondition(currentDay))
	for i, user := range minutes {
		if user.userId == id {
			// rank and time
			rankString := "Rang #" + fmt.Sprint(i+1) + " mit " + formatTime(user.minutes)
			if member == author {
				rankString = (*member).Mention() + ", du bist " + rankString
			} else {
				rankString = effectiveName(member) + " ist " + rankString
			}

			type ChannelMinutes struct {
				channel int16
				minutes uint32
			}

			// top channels
			channels := make([]ChannelMinutes, 0, len(user.channels))
			for channel, minutes := range user.channels {
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
				channel, err := dc.Channel(longChannelId(channels[i].channel))
				name := "[Gelöschter Kanal]"
				if err == nil {
					name = channel.Name
				}
				rankString += "\n" + digitEmote(i+1) + ": " + name + ": " + formatTime(channels[i].minutes)
			}
			dc.ChannelMessageSend(channel, rankString)
			return
		}
	}
	dc.ChannelMessageSend(channel, "Keine aufgezeichnete Sprachzeit für den Nutzer "+effectiveName(member)+" gefunden")
}

func hoursCommand(args string, channel string, author *discordgo.Member) {
	min := func(a, b int16) int16 {
		if a < b {
			return a
		} else {
			return b
		}
	}

	ok, dateCondition, member := parseUserOrTime(args, channel, author)
	if !ok {
		return
	} // error messages handled by util function, just return

	const SECTION_SIZE = 5
	const SECTIONS = (24 * 60) / SECTION_SIZE

	shortUser := shortUserId(member.User.ID)
	minutesSum := uint32(0)
	minutesBySections := make([]uint32, SECTIONS)

	dayCount := 0
	for date, day := range dayData {
		if !dateCondition(date) {
			continue
		}
		dayCount++
		for _, sessions := range day.channels {
			for _, session := range sessions {
				if session.userID != shortUser {
					continue
				}
				// session from requested user
				currentMinute := session.dayMinute
				remaining := session.minutes
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
		dc.ChannelMessageSend(channel, "Keine Daten gefunden!")
		return
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

func channelsCommand(args string, channel string, author *discordgo.Member) {
	ok, dateCondition, member := parseUserOrTime(args, channel, author)
	if !ok {
		return
	} // error messages handled by util function, just return

	shortUser := shortUserId(member.User.ID)

	stats, _ := calculateUserMinutes(dateCondition, dateCondition(currentDay))

	for _, userStats := range stats {
		if userStats.userId != shortUser {
			continue
		}

		values := make([]chart.Value, len(userStats.channels))
		for id, minutes := range userStats.channels {
			name := "[Gelöschter Kanal]"
			// don't label channels with small pieces
			if (userStats.minutes / minutes) < 50 {
				if channel, ok := dc.Channel(longChannelId(id)); ok == nil {
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
			log.Fatal("Error while creating diagram: ", err)
		}
		dc.ChannelFileSend(channel, "channels.png", bytes.NewReader(buffer.Bytes()))
		return
	}
	dc.ChannelMessageSend(channel, "Keine Daten gefunden!")
}

func sessionHandler(args string, channel string, author *discordgo.Member) {
	if session, ok := sessions[shortUserId(author.User.ID)]; ok {
		voiceChannel, _ := dc.State.Channel(longChannelId(session.channelID))

		dc.ChannelMessageSend(channel, fmt.Sprintf("Du bist seit %v Minuten im Kanal %v", session.session.minutes, voiceChannel.Name))
	} else {
		dc.ChannelMessageSend(channel, "Du bist momentan in keinem Sprachkanal")
	}
}

func mateHandler(args string, channel string, author *discordgo.Member) {
	if args == "" {
		dc.ChannelMessageSend(channel, "Bitte gib einen Nutzernamen an: **!mate {Nutzer}**")
		return
	}
	mateMember := parseMember(args)
	if mateMember == nil {
		dc.ChannelMessageSend(channel, "Der angegebene Nutzer wurde nicht gefunden!")
		return
	}
	mateId := shortUserId(mateMember.User.ID)
	authorId := shortUserId(author.User.ID)
	mates, _ := timeWithMates(authorId, dateAllTimeCondition)
	for i, mate := range mates {
		if mate.userId == mateId {
			dc.ChannelMessageSend(channel, fmt.Sprintf(
				"%v, deine überlappende Zeit mit %v beträgt %v (Platz %v)",
				author.Mention(), effectiveName(mateMember), formatTime(mate.minutes), i+1))
		}
	}
}

func matesHandler(args string, channel string, author *discordgo.Member) {
	const matesToShow = 9

	authorId := shortUserId(author.User.ID)
	mates, allMatesTime := timeWithMates(authorId, dateAllTimeCondition)
	text := author.Mention() + ", deine Top-Freunde sind:\n"
	values := []chart.Value{}
	topMatesTime := uint32(0)
	listCount := matesToShow
	if len(mates) < matesToShow {
		if len(mates) == 0 {
			dc.ChannelMessageSend(channel, "Du hast keine Freunde "+author.Mention()+" :cry:")
			return
		}
		listCount = len(mates)
	}
	for i := 0; i < listCount; i++ {
		mateMember, _ := dc.State.Member(config.GuildId, longUserId(mates[i].userId))
		mateName := effectiveName(mateMember)
		values = append(values, chart.Value{
			Value: float64(mates[i].minutes) / float64(allMatesTime),
			Label: mateName,
		})
		topMatesTime += mates[i].minutes
		text += fmt.Sprintf("%v: %v (%v)\n", digitEmote(i+1), mateName, formatTime(mates[i].minutes))
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
		log.Fatal("Error while creating diagram: ", err)
	}
	dc.ChannelFileSend(channel, "mates.png", bytes.NewReader(buffer.Bytes()))
	dc.ChannelMessageSend(channel, text)
}

func stonksHandler(args string, channel string, author *discordgo.Member) {
	ok, dateCondition, member := parseUserOrTime(args, channel, author)
	if !ok {
		return
	} // error messages handled by util function, just return

	memberId := shortUserId(member.User.ID)

	dates := make([]Date, 0, len(dayData))

	for date := range dayData {
		dates = append(dates, date)
	}

	sort.Slice(dates, func(i, j int) bool {
		return dateIsSmaller(dates[i], dates[j])
	})

	starting := true

	var xValues []time.Time
	var yValues []float64
	var allTimeMinutes uint32
	for _, date := range dates {
		if !dateCondition(date) { continue }
		var minutes uint32
		for _, channel := range dayData[date].channels {
			for _, session := range channel {
				if session.userID == memberId {
					minutes += uint32(session.minutes)
				}
			}
		}
		if minutes > 0 || !starting {
			starting = false
			allTimeMinutes += minutes
			xValues = append(xValues, time.Date(int(date.year), time.Month(date.month), int(date.day), 0, 0, 0, 0, time.Local))
			yValues = append(yValues, float64(allTimeMinutes))
		}
	}
	if len(xValues) < 2 {
		dc.ChannelMessageSend(channel, "Nicht genug Daten gefunden!")
		return
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
		Title:        "Stonks von " + effectiveName(member),
		XAxis: chart.XAxis{
			Name:           "Datum",
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			Name: "Stunden",
			ValueFormatter: func(v interface{}) string {
				if typed, isTyped := v.(float64); isTyped {
					return formatTime(uint32(typed))
				}
				return "error"
			},
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
		log.Fatal("Error while creating diagram: ", err)
	}
	dc.ChannelFileSend(channel, "stonks.png", bytes.NewReader(buffer.Bytes()))
}
