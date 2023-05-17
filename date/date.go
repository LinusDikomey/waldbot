package date

import (
    "time"
)

var CurrentDay Date

type DateCondition = func(Date) bool

type Date struct {
	Day   uint16
	Month uint8
	Year  uint16
}

func New(day uint16, month uint8, year uint16) Date {
	return Date{Day: day, Month: month, Year: year}
}

func FromTime(t time.Time) Date {
	return Date {
        Day: uint16(t.Day()),
        Month: uint8(t.Month()),
        Year: uint16(t.Year()),
    }
}


type sortDates []Date

func (s sortDates) Len() int {
    return len(s)
}
func (s sortDates) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s sortDates) Less(i, j int) bool {
    if s[i].Year == s[j].Year {
		if s[i].Month == s[j].Month {
			return s[i].Day < s[j].Day
		} else {
			return s[i].Month < s[j].Month
		}
	} else {
		return s[i].Year < s[j].Year
	}
}

func IsSmaller(a, b Date) bool {
    if a.Year == b.Year {
		if a.Month == b.Month {
			return a.Day < b.Day
		} else {
			return a.Month < b.Month
		}
	} else {
		return a.Year < b.Year
	}
}

func DayMinute(t time.Time) int16 {
	return int16(t.Hour() * 60 + t.Minute())
}
