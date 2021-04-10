package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/golang/freetype/truetype"
)


var (
	normalFont *truetype.Font
	boldFont *truetype.Font
)


func formatTime(minutes uint32) string {
	hours := minutes / 60
	hourMinutes := minutes % 60
	minutesString := fmt.Sprint(hourMinutes)
	if len(minutesString) == 1 {
		minutesString = "0" + minutesString
	}
	return fmt.Sprintf("%v:%vh", hours, minutesString)
}

func effectiveName(member *discordgo.Member) string {
	var name string
	if member == nil {
		name = "[Unbekannter Nutzer]"
	} else if member.Nick != "" {
		name = member.Nick
	} else {
		name = member.User.Username
	}
	name = strings.ReplaceAll(name, "https://", "bruh://")
	name = strings.ReplaceAll(name, "http://", "bruh://")
	return name
}

func digitEmote(digit int) string {
	switch digit {
	case 0:
		return ":zero:"
	case 1:
		return ":one:"
	case 2:
		return ":two:"
	case 3:
		return ":three:"
	case 4:
		return ":four:"
	case 5:
		return ":five:"
	case 6:
		return ":six:"
	case 7:
		return ":seven:"
	case 8:
		return ":eight:"
	case 9:
		return ":nine:"
	default:
		return ":question:"
	}
}

func loadFonts() {
	normalFile, _ := os.Open("./data/fonts/whitneylight.ttf")
	//normalFile, _ := os.Open("./data/fonts/whitneymedium.ttf")
	boldFile, _ := os.Open("./data/fonts/whitneybold.ttf")
	normalBytes, _ := ioutil.ReadAll(normalFile)
	boldBytes, _ := ioutil.ReadAll(boldFile)
	normalFont, _ = truetype.Parse(normalBytes)
	boldFont, _ = truetype.Parse(boldBytes)
}

func charWidth(font *truetype.Font, c rune) int {
	hMetric := font.HMetric(1000 << 6, font.Index(c))
	return hMetric.AdvanceWidth.Ceil()
}

func stringWidth(font *truetype.Font, str string) int {
	width := 0
	for c := range str {
		width += charWidth(font, rune(c))
	}
	return width
}

func parseMember(str string) *discordgo.Member {
	if str == "" {
		return nil
	}
	members, _ := dc.GuildMembers(config.GuildId, "0", 1000)
	for _, member := range(members) {
		if member.Nick == str ||
		   member.User.Username == str ||
		   member.User.Username + "#" + member.User.Discriminator == str {
			return member
		}
	}
	for _, member := range(members) {
		if strings.ToLower(member.Nick) == strings.ToLower(str) ||
		   strings.ToLower(member.User.Username) == strings.ToLower(str) ||
		   strings.ToLower(member.User.Username + "#" + member.User.Discriminator) == strings.ToLower(str) {
			return member
		}
	}
	return nil
}

func parseDate(str string) *Date {
	split := strings.Split(str, ".")
	if len(split) != 3 {
		return nil
	}
	day, err1 := strconv.Atoi(split[0])
	month, err2 := strconv.Atoi(split[1])
	year, err3 := strconv.Atoi(split[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}
	if day < 1 || day > 31 || month < 1 || month > 12 || year < 1 || year > 65000 {
		return nil
	}
	date := newDate(uint16(day), uint8(month), uint16(year))
	return &date
}

const (
	parseSuccess = iota
	parseInvalid = iota
	parseNone = iota
)

// Tries to parse the args string as any date range. 
// Will always return back the default condition and wether it successfully parsed a date
func parseDateCondition(args string, defaultCondition DateCondition) (DateCondition, uint8) {
	condition := defaultCondition
	if args != "" {
		if args == "daily" {
			return dateDailyCondition, parseSuccess
		} else if args == "weekly" {
			return dateWeeklyCondition, parseSuccess
		} else if args == "monthly" {
			return dateMonthlyCondition, parseSuccess
		} else if args == "yearly" {
			return dateYearlyCondition, parseSuccess
		} else if args == "all" || args == "allTime" {
			return dateAllTimeCondition, parseSuccess	
		} else {
			dateRangeSplit := strings.Split(args, "-")
			dateCount := len(dateRangeSplit)
			isDate := false
			if dateCount == 1 || dateCount == 2 {
				startDate := parseDate(dateRangeSplit[0])
				endDate := startDate
				
				if len(strings.Split(dateRangeSplit[0], ".")) == 3 {
					if dateCount == 2 && len(strings.Split(dateRangeSplit[1], ".")) == 3 {
						endDate = parseDate(dateRangeSplit[1])
					}
					if startDate == nil || endDate == nil {
						return condition, parseInvalid
					}

					isDate = true
				} else if len(dateRangeSplit) == 1 && len(strings.Split(dateRangeSplit[0], ".")) == 2 {
					// month.year (take whole month as range)
					split := strings.Split(dateRangeSplit[0], ".")
					month, err1 := strconv.Atoi(split[0])
					year, err2 := strconv.Atoi(split[1])
					if err1 == nil || err2 == nil {
						if month < 1 || month > 12 || year < 0 || year > 65000 {
							return condition, parseInvalid
						}
						startDate = &Date {day: 1, month: uint8(month), year: uint16(year)}
						endDate = &Date {day: 31, month: uint8(month), year: uint16(year)}
						isDate = true
					}
				}
				if isDate {
					condition = func(day Date) bool {
						return !dateIsSmaller(day, *startDate) && !dateIsSmaller(*endDate, day)
					}
					return condition, parseSuccess
				} else {
					return condition, parseNone
				}
			} else {
				return condition, parseNone
			}
		}
	} else {
		return condition, parseNone
	}
}

func parseMemberOrDateCondition(args string, channel string, member *discordgo.Member) (bool, func(Date) bool, *discordgo.Member) {
	condition, success := parseDateCondition(args, dateAllTimeCondition)
	if success == parseInvalid {
		dc.ChannelMessageSend(channel, "Invalides Datum angegeben!")
		return false, nil, nil
	}
	if args != "" {	
		if success == parseNone {
			member = parseMember(args)
			if member == nil {
				dc.ChannelMessageSend(channel, "Der angegebene Nutzer wurde nicht gefunden!")
				return false, nil, nil
			}
		}
	}
	return true, condition, member
}