package vacation

import (
	"time"
)

const (
	// VacationLimitDays максимальное количество дней для выбора даты отпуска
	VacationLimitDays = 21
)

// GenerateVacationDates генерирует список дат для выбора даты выхода на работу
func (s *ServiceVacation) generateVacationDates() []string {
	var dates []string
	tomorrow := s.getTomorrow()

	for i := 0; i < VacationLimitDays; i++ {
		date := tomorrow.AddDate(0, 0, i)
		dates = append(dates, date.Format(DateFormatLayout))
	}

	return dates
}

// IsDateFormat проверяет, является ли текст датой в валидном формате
func (s *ServiceVacation) isDateFormat(text string) bool {
	_, err := time.Parse(DateFormatLayout, text)
	return err == nil
}

// IsValidVacationDate проверяет, что дата не раньше завтрашнего дня и не позже чем через 3 недели
func (s *ServiceVacation) isValidVacationDate(date time.Time) bool {
	tomorrow := s.getTomorrow()
	maxDate := s.getMaxVacationDate()

	return !date.Before(tomorrow) && !date.After(maxDate)
}

// GetTomorrow возвращает дату на завтра (начало дня)
func (s *ServiceVacation) getTomorrow() time.Time {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
}

// GetMaxVacationDate возвращает макgсимально допустимую дату для выхода из отпуска
func (s *ServiceVacation) getMaxVacationDate() time.Time {
	maxDate := time.Now().AddDate(0, 0, VacationLimitDays+1) // VacationLimitDays дней + 1 сегодня
	return time.Date(maxDate.Year(), maxDate.Month(), maxDate.Day(), 0, 0, 0, 0, maxDate.Location())
}
