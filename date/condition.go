package date

import (
	"time"
)

func DailyCondition(day Date) bool {
	return day == CurrentDay
}

func WeeklyCondition(day Date) bool {
	now := time.Now()
 
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
    mondayDate := Date {
        Day: uint16(weekStartDate.Day()),
        Month: uint8(weekStartDate.Month()),
        Year: uint16(weekStartDate.Year()),
    }
	return !IsSmaller(day, mondayDate)
}

func MonthlyCondition(day Date) bool {
	firstOfMonth := Date { Day: 1, Month: CurrentDay.Month, Year: CurrentDay.Year }
	return !IsSmaller(day, firstOfMonth)
}

func YearlyCondition(day Date) bool {
	firstOfYear := Date { Day: 1, Month: 1, Year: CurrentDay.Year }
	return !IsSmaller(day, firstOfYear)
}

func AllTimeCondition(day Date) bool {
	return true
}
