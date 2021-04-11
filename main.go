package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/LinusDikomey/waldbot/oauth"
	"github.com/bwmarrin/discordgo"
)

type ActiveSession struct {
	session VoiceSession
	channelID ChannelId
}

var (
	dayData map[Date]Day = map[Date]Day {}
	sessions map[UserId]ActiveSession = map[UserId]ActiveSession {}
	dc *discordgo.Session
	currentDay Date
	targetWidth int
	guild *discordgo.Guild
	oauthClientId string
	oauthClientSecret string
)

func getEnv(name string) string {
	env := os.Getenv(name)
	if env == "" {
		fmt.Println("Please set the environment variable '" + name + "'!")
		os.Exit(-1)
	}
	return env
}

func main() {
	fmt.Println("Starting Waldbot...")
	loadData()
	loadConfig()
	loadFonts()
	targetWidth = stringWidth(boldFont, "__Osabama gone Oase") + stringWidth(normalFont, ": 10000:00h")

	oauth.ClientId = getEnv("WALDBOT_CLIENTID")
	oauth.ClientSecret = getEnv("WALDBOT_CLIENTSECRET")
	token := getEnv("WALDBOT_TOKEN")

	readDays("./data/days/")
	if token == "" {
		fmt.Println("Please set the environment variable 'WALDBOT_TOKEN'")
	}

	if token == "" {
		fmt.Println("Could not find the bot token, please set the environment variable 'WALDBOT_TOKEN' to the application token!")
		return
	}
	var err error
	dc, err = discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error while starting discord session, ", err)
	}

	// handlers
	dc.AddHandler(onMessageCreate)
	dc.AddHandler(onChannelConnect)
	dc.AddHandler(onChannelDisconnect)
	dc.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})
	dc.Identify.Intents = discordgo.IntentsAll
	dc.StateEnabled = true

	// Open a websocket connection to Discord and begin listening.
	err = dc.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	/*for _, v := range slashCommands {
		fmt.Println("Adding command:", v.Name, ", stateId:", dc.State.User.ID, ", guildId:", config.GuildId, ", v:", v)
		_, err := dc.ApplicationCommandCreate(dc.State.User.ID, config.GuildId, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}*/

	guild, err = dc.Guild(config.GuildId)
	if err != nil {
		log.Fatal("Could not find guild specified in config!")
	}

	addWebHandlers()
	fmt.Println("Bot is now running!")

	currentDay = dateFromTime(time.Now())
	
	updateRankings()
	
	mainLoop()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("Stopping...")
	dc.Close()
	finishSessions()
	saveFiles()
}

func dayMinute(t time.Time) int16 {
	return int16(t.Hour() * 60 + t.Minute())
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	content, _ := m.Message.ContentWithMoreMentionsReplaced(dc)
	if len(content) == 0 {
		return
	}
	if content[0] == '!' {
		command := content[1:]
		split := strings.SplitN(command, " ", 2)
		prefix := split[0]
		var args string
		if len(split) == 2 {
			args = split[1]
		} else {
			args = ""
		}
		for _, command := range commands {
			if prefix == command.prefix {
				m.Member.User = m.Author
				command.handler(args, m.ChannelID, m.Member)
			}
		}
	}
}

func mainLoop() {
	currentTime := time.Now()
	currentTime.Second()
	quit := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(60 - currentTime.Second()) * time.Second)
		minuteUpdate()
		ticker := time.NewTicker(60 * time.Second)
    	for {
			select {
			case <- ticker.C:
            	minuteUpdate()
        		case <- quit:
            		ticker.Stop()
            		return
        	}
    	}
 	}()
}

func minuteUpdate() {
	currentTime := time.Now()
	date := dateFromTime(currentTime)
	
	if _, ok := dayData[date]; !ok {
		fmt.Println("New day started!")
		dayData[date] = Day {channels: map[int16][]VoiceSession{}}
		dayUpdate(currentDay, date)
		currentDay = date
	}
	dayMinute := dayMinute(currentTime)
	
	if dayMinute % 30 == 0 {
		saveFiles()
	}

	members, err := dc.GuildMembers(config.GuildId, "", 1000)
	if err != nil {
		panic("Could not find Guild by id specified in config: " + fmt.Sprint(err))
	}
	for _, member := range members {
		if member.User.Bot { continue }
		state, _ := dc.State.VoiceState(config.GuildId, member.User.ID)
		shortID := shortUserId(member.User.ID)
		if state != nil && !state.Mute && !state.SelfMute && !state.Deaf && !state.SelfDeaf && !state.Suppress {
			// user is active in a voice chat
			
			currentChannelId := shortChannelId(state.ChannelID)
			if active, ok := sessions[shortID]; ok {
				if active.channelID == currentChannelId {
					// continue existing session
					active.session.minutes++
					sessions[shortID] = active
				} else {
					//channel changed, finish existing and start new session
					dayData[date].channels[active.channelID] = append(dayData[date].channels[active.channelID], active.session)

					newChannelId := currentChannelId
					sessions[shortID] = ActiveSession{session: VoiceSession{ dayMinute: dayMinute, userID: shortID, minutes: 1 }, channelID: newChannelId}		
				}
			} else {
				//new session
				shortChannelID := currentChannelId
				sessions[shortID] = ActiveSession{session: VoiceSession{ dayMinute: dayMinute, userID: shortID, minutes: 1 }, channelID: shortChannelID}
			}
		} else {
			// user is not in a voice chat
			if finished, ok := sessions[shortID]; ok {
				// session just ended, append to dayData
				dayData[date].channels[finished.channelID] = append(dayData[date].channels[finished.channelID], finished.session)
				delete(sessions, shortID) 
			}
		}
	}
	updateRankings()
}

func dayUpdate(previousDay Date, newDay Date) {
	// finish sessions from previous day
	finishSessions()
	saveFiles()
}

func finishSessions() {
	for _, active := range(sessions) {
		dayData[currentDay].channels[active.channelID] = append(dayData[currentDay].channels[active.channelID], active.session)
	}
	sessions = make(map[int16]ActiveSession)
}

func dateDailyCondition(day Date) bool {
	return day == currentDay
}

func dateWeeklyCondition(day Date) bool {
	now := time.Now()
 
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	mondayDate := newDate(uint16(weekStartDate.Day()), uint8(weekStartDate.Month()), uint16(weekStartDate.Year()))
	return !dateIsSmaller(day, mondayDate)
}
type DateCondition = func(Date) bool

func dateMonthlyCondition(day Date) bool {
	firstOfMonth := Date {day: 1, month: currentDay.month, year: currentDay.year }
	return !dateIsSmaller(day, firstOfMonth)
}

func dateYearlyCondition(day Date) bool {
	firstOfYear := Date {day: 1, month: 1, year: currentDay.year}
	return !dateIsSmaller(day, firstOfYear)
}

func dateAllTimeCondition(day Date) bool {
	return true
}

func updateRankings() {
	daily, _ := calculateUserMinutes(dateDailyCondition, true)
	dc.ChannelMessageEdit(config.StatsChannelId, fmt.Sprint(data.ServerdatenDailyMessageId), createRankingsMessage("**Heutige Sprachchatzeiten**", daily, 1))
	weekly, weeklyDays := calculateUserMinutes(dateWeeklyCondition, true)
	dc.ChannelMessageEdit(config.StatsChannelId, fmt.Sprint(data.ServerdatenWeeklyMessageId), createRankingsMessage("**Wöchentliche Sprachchatzeiten**", weekly, weeklyDays))
	allTime, allTimeDays := calculateUserMinutes(dateAllTimeCondition, true)
	dc.ChannelMessageEdit(config.StatsChannelId, fmt.Sprint(data.ServerdatenMessageId), createRankingsMessage("**Gesamte Sprachchatzeiten**", allTime, allTimeDays))

	// api stats
	apiStats = ApiStats {
		Ranking: make([]ApiUserStats, len(allTime)),
	}
	for i, user := range allTime {
		name := ""
		if member, err := dc.State.Member(config.GuildId, longUserId(user.userId)); err == nil {
			name = effectiveName(member)
		}

		apiStats.Ranking[i] = ApiUserStats {
			Username: name,
			Minutes: user.minutes,
		}
	}
}

func createRankingsMessage(header string, stats []UserStats, dayCount int) string {
	maxHeaderLength := len(header) + 30
	allMinutes := uint32(0)
	rankings := ""
	for n, entry := range stats {
		rank := n + 1
		allMinutes += entry.minutes
		if rank <= 15 {
			linePrefix := ""
			if rank < 10 {
				linePrefix += ":zero:" + digitEmote(rank)
			} else {
				linePrefix += ":one:" + digitEmote(rank - 10)
			}
			linePrefix += ": "
			member, err := dc.State.Member(config.GuildId, longUserId(entry.userId))
			var name string
			if err != nil {
				name = "[Unbekannt]"
			} else {
				name = effectiveName(member)
			}
			if len(name) > 18 {
				name = name[:15] + "..."
			}
			time := formatTime(entry.minutes)
			line := linePrefix + "**" + name + "**: " + time
			width := stringWidth(boldFont, name) + stringWidth(normalFont, ": " + time)
			spaceLength := charWidth(normalFont, ' ')
			spacesToAdd := float64(targetWidth - width) / float64(spaceLength)
			for i := 0; i < int(math.Round(spacesToAdd)); i++ {
				line += " "
			}
			// Topkanal
			var topChannel int16
			topMinutes := uint32(0)
			for channelID, minutes := range entry.channels {
				if minutes > topMinutes {
					topChannel = channelID
					topMinutes = minutes
				}
			}
			channelName := "[Gelöschter Kanal]"
			if ch, err := dc.State.Channel(longChannelId(topChannel)); err == nil {
				channelName = ch.Name
			}
			line += "Topkanal: " + channelName + " (" + formatTime(topMinutes) + ")\n"
			if len(rankings) + len(line) + maxHeaderLength >= 2000 {
				continue
			}
			rankings += line
		}
	}
	if dayCount > 1 {
		header += " (" + fmt.Sprint(dayCount) + " Tage, " + formatTime(allMinutes) + ")"
	} else {
		header += " (" + formatTime(allMinutes) + ")"
	}
	return header + "\n\n" + rankings + "\n" + "‎" // <--- there is a left-to-right mark here
}

func saveFiles() {
	saveData()
	if currentDayData, ok := dayData[currentDay]; ok {
		file, err := os.Create(fmt.Sprint("./data/days/", currentDay.day, "-", currentDay.month, "-", currentDay.year, ".day"))
		defer file.Close()
		if err != nil {
			log.Fatal("Could not save day data to file: ", err)
		}
		currentDayData.save(file)
	}
}

type UserStats struct {
	userId int16
	minutes uint32
	channels map[ChannelId]uint32 
}

type SortUserStats []UserStats

func (s SortUserStats) Len() int {
	return len(s)
}
func (s SortUserStats) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortUserStats) Less(i, j int) bool {
	return s[i].minutes < s[j].minutes
}

// returns sorted list of UserStats
func calculateUserMinutes(dayCondition func(Date) bool, includeActive bool) ([]UserStats, int) {
	dayCount := 0
	users := map[int16]UserStats {}
	addSession := func(channel int16, session VoiceSession) {
		// get or create stats for user
		var userStats UserStats
		if stats, ok := users[session.userID]; ok {
			userStats = stats
		} else {
			userStats = UserStats{userId: session.userID, minutes: 0, channels: map[ChannelId]uint32 {}}
		}
		// increase overall minutes
		userStats.minutes += uint32(session.minutes)
		// get or initialize stats for channel
		channelMinutes := uint32(0)
		if minutes, ok := userStats.channels[channel]; ok {
			channelMinutes = minutes
		}
		// increase minutes for channel
		userStats.channels[channel] = channelMinutes + uint32(session.minutes)
		// put userStats back into users map
		users[session.userID] = userStats
	}

	for date, day := range dayData {
		if dayCondition(date) {
			dayCount++
			for channel, sessions := range day.channels {
				for _, session := range sessions {
					addSession(channel, session)
				}
			}
		}
	}
	if includeActive {
		for _, active := range sessions {
			addSession(active.channelID, active.session)
		}
	}
	// put values into list and sort
	list := make([]UserStats, 0, len(users))
	for _, val := range users {
		list = append(list, val)
	}
	f := func (n, n1 int) bool {
		return list[n].minutes > list[n1].minutes
	}
	sort.Slice(list, f)
	return list, dayCount
}

func timeWithMates(user int16, dateCondition DateCondition) ([]UserStats, uint32) {
	min := func(a, b int32) int32 {
		if a < b { return a } else { return b }
	}
	mates := map[int16]UserStats {}
	allTime := uint32(0)
	for day, data := range dayData {
		if !dateCondition(day) { continue }
		for _, sessions := range data.channels {
			for _, session := range sessions {
				if session.userID != user { continue }
				for _, otherSession := range sessions {
					if otherSession.userID == user { continue }
					var minutes int32
					a := int32(session.dayMinute)
					b := a + int32(session.minutes)
					c := int32(otherSession.dayMinute)
					d := c + int32(otherSession.minutes)
					if a < c {
						minutes = min(b-c, d-c)
					} else {
						minutes = min(d-a, b-a)
					}
					if minutes > 0 {
						var stats UserStats
						if found, ok := mates[otherSession.userID]; ok {
							stats = found
						} else {
							stats = UserStats {userId: otherSession.userID, minutes: 0} 
						}
						stats.minutes += uint32(minutes)
						mates[otherSession.userID] = stats

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
		return slice[i].minutes > slice[j].minutes
	})
	return slice, allTime
}