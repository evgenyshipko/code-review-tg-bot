package constants

const (
	Start            = "start"
	TakeVacation     = "take_vacation"
	ReturnToWork     = "return_to_work"
	SendOnVacation   = "send_on_vacation"
	VacationsList    = "vacations_list"
	TestTakeVacation = "test_take_vacation"
)

type ButtonText struct {
	TakeVacation       string
	ReturnFromVacation string
	ChangeVacation     string
	ReturnToWork       string
	Cancel             string
	VacationsList      string
	SendOnVacation     string
}

var ButtonTextConstants = ButtonText{
	TakeVacation:       "🏃 Оформить отпуск",
	ReturnFromVacation: "🔄 Вернуть из отпуска:",
	ChangeVacation:     "📅 Изменить отпуск:",
	ReturnToWork:       "💼 Выход на работу",
	Cancel:             "❌ Закрыть",
	VacationsList:      "📋 Список отпусков",
	SendOnVacation:     "➕ Отправить в отпуск",
}

var KeyboardToRoleMapping = map[string][]UserRole{
	ButtonTextConstants.ReturnToWork:       []UserRole{Reviewer},
	ButtonTextConstants.TakeVacation:       []UserRole{Reviewer},
	ButtonTextConstants.ReturnFromVacation: []UserRole{Admin},
	ButtonTextConstants.SendOnVacation:     []UserRole{Admin},
	ButtonTextConstants.VacationsList:      []UserRole{Admin},
	ButtonTextConstants.Cancel:             []UserRole{Admin, Tester, Reviewer},
}

type UserRole string

const (
	Tester   UserRole = "Tester"
	Reviewer UserRole = "Reviewer"
	Admin    UserRole = "Admin"
)
