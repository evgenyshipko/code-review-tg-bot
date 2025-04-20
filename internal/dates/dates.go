package dates

import (
	"code-review-tg-bot/internal/constants"
	"time"
)

const (
	VacationLimitDays = 21
)

func GenerateVacationDates() []string {
	var dates []string
	tomorrow := GetTomorrow()

	for i := 0; i < VacationLimitDays; i++ {
		date := tomorrow.AddDate(0, 0, i)
		dates = append(dates, date.Format(constants.DateFormatLayout))
	}

	return dates
}

func IsValidVacationDate(date time.Time) bool {
	tomorrow := GetTomorrow()
	maxDate := GetMaxVacationDate()

	return !date.Before(tomorrow) && !date.After(maxDate)
}

func GetTomorrow() time.Time {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
}

func GetMaxVacationDate() time.Time {
	maxDate := time.Now().AddDate(0, 0, VacationLimitDays+1) // VacationLimitDays дней + 1 сегодня
	return time.Date(maxDate.Year(), maxDate.Month(), maxDate.Day(), 0, 0, 0, 0, maxDate.Location())
}
