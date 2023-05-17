package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/golang/freetype/truetype"

    "waldbot/date"
)


var (
	NormalFont *truetype.Font
	BoldFont *truetype.Font
)


func FormatTime(minutes uint32) string {
	hours := minutes / 60
	hourMinutes := minutes % 60
	minutesString := fmt.Sprint(hourMinutes)
	if len(minutesString) == 1 {
		minutesString = "0" + minutesString
	}
	return fmt.Sprintf("%v:%vh", hours, minutesString)
}

func EffectiveName(member *discordgo.Member) string {
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

func DigitEmote(digit int) string {
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

func LoadFonts() {
	normalFile, _ := os.Open("./data/fonts/whitneylight.ttf")
	//normalFile, _ := os.Open("./data/fonts/whitneymedium.ttf")
	boldFile, _ := os.Open("./data/fonts/whitneybold.ttf")
	normalBytes, _ := ioutil.ReadAll(normalFile)
	boldBytes, _ := ioutil.ReadAll(boldFile)
	NormalFont, _ = truetype.Parse(normalBytes)
	BoldFont, _ = truetype.Parse(boldBytes)
}

func CharWidth(font *truetype.Font, c rune) int {
	hMetric := font.HMetric(1000 << 6, font.Index(c))
	return hMetric.AdvanceWidth.Ceil()
}

func StringWidth(font *truetype.Font, str string) int {
	width := 0
	for c := range str {
		width += CharWidth(font, rune(c))
	}
	return width
}

func ParseMember(guildID string, str string) *discordgo.Member {
	if str == "" {
		return nil
	}
	members, _ := Dc.GuildMembers(guildID, "0", 1000)
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

func parseDate(str string) *date.Date {
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
	date := date.New(uint16(day), uint8(month), uint16(year))
	return &date
}

const (
	ParseSuccess = iota
	ParseInvalid = iota
	ParseNone = iota
)

// Tries to parse the args string as any date range. 
// Will always return back the default condition and wether it successfully parsed a date
func ParseDateCondition(args string, defaultCondition date.DateCondition) (date.DateCondition, uint8) {
	condition := defaultCondition
	if args != "" {
		if args == "daily" {
			return date.DailyCondition, ParseSuccess
		} else if args == "weekly" {
			return date.WeeklyCondition, ParseSuccess
		} else if args == "monthly" {
			return date.MonthlyCondition, ParseSuccess
		} else if args == "yearly" {
			return date.YearlyCondition, ParseSuccess
		} else if args == "all" || args == "allTime" {
			return date.AllTimeCondition, ParseSuccess	
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
						return condition, ParseInvalid
					}

					isDate = true
				} else if len(dateRangeSplit) == 1 && len(strings.Split(dateRangeSplit[0], ".")) == 2 {
					// month.year (take whole month as range)
					split := strings.Split(dateRangeSplit[0], ".")
					month, err1 := strconv.Atoi(split[0])
					year, err2 := strconv.Atoi(split[1])
					if err1 == nil || err2 == nil {
						if month < 1 || month > 12 || year < 0 || year > 65000 {
							return condition, ParseInvalid
						}
						startDate = &date.Date {Day: 1, Month: uint8(month), Year: uint16(year)}
						endDate = &date.Date {Day: 31, Month: uint8(month), Year: uint16(year)}
						isDate = true
					}
				}
				if isDate {
					condition = func(day date.Date) bool {
						return !date.IsSmaller(day, *startDate) && !date.IsSmaller(*endDate, day)
					}
					return condition, ParseSuccess
				} else {
					return condition, ParseNone
				}
			} else {
				return condition, ParseNone
			}
		}
	} else {
		return condition, ParseNone
	}
}

func parseMemberOrDateCondition(guildID string, args string, channel string, member *discordgo.Member) (bool, func(date.Date) bool, *discordgo.Member) {
	condition, success := ParseDateCondition(args, date.AllTimeCondition)
	if success == ParseInvalid {
		Dc.ChannelMessageSend(channel, "Invalides Datum angegeben!")
		return false, nil, nil
	}
	if args != "" {	
		if success == ParseNone {
			member = ParseMember(guildID, args)
			if member == nil {
				Dc.ChannelMessageSend(channel, "Der angegebene Nutzer wurde nicht gefunden!")
				return false, nil, nil
			}
		}
	}
	return true, condition, member
}


