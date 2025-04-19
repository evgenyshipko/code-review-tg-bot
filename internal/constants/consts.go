package constants

const (
	Start           = "start"
	Rest            = "rest"
	Work            = "work"
	Vacations_start = "vacations_start"
	Vacations       = "vacations"
	Test_vacation   = "test_vacation"
)

type ButtonText struct {
	TakeVacation       string
	ReturnFromVacation string
	ChangeVacation     string
	ReturnToWork       string
	Cancel             string
	VacationsList      string
	VacationsStart     string
}

var ButtonTextConstants = ButtonText{
	TakeVacation:       "🏃 Оформить отпуск",
	ReturnFromVacation: "🔄 Вернуть из отпуска:",
	ChangeVacation:     "📅 Изменить отпуск:",
	ReturnToWork:       "💼 Выход на работу",
	Cancel:             "❌ Закрыть",
	VacationsList:      "📋 Список отпусков",
	VacationsStart:     "➕ Отправить в отпуск",
}

var KeyboardToRoleMapping = map[string][]UserRole{
	ButtonTextConstants.ReturnToWork:       []UserRole{Reviewer},
	ButtonTextConstants.TakeVacation:       []UserRole{Reviewer},
	ButtonTextConstants.ReturnFromVacation: []UserRole{Admin},
	ButtonTextConstants.VacationsStart:     []UserRole{Admin},
	ButtonTextConstants.VacationsList:      []UserRole{Admin},
	ButtonTextConstants.Cancel:             []UserRole{Admin, Tester, Reviewer},
}

type UserRole string

const (
	Tester   UserRole = "Tester"
	Reviewer UserRole = "Reviewer"
	Admin    UserRole = "Admin"
)
