package constants

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type BotCommand string

const (
	Start            BotCommand = "start"
	TakeVacation     BotCommand = "take_vacation"
	ReturnToWork     BotCommand = "return_to_work"
	SendToVacation   BotCommand = "send_on_vacation"
	VacationsList    BotCommand = "vacations_list"
	TestTakeVacation BotCommand = "test_take_vacation"
)

type ButtonText struct {
	ReturnFromVacation string
	ChangeVacation     string
	Cancel             string
}

var ButtonTextConstants = ButtonText{
	ReturnFromVacation: "🔄 Вернуть из отпуска:",
	ChangeVacation:     "📅 Изменить отпуск:",
	Cancel:             "❌ Закрыть",
}

type UserRole string

const (
	Tester   UserRole = "Tester"
	Reviewer UserRole = "Reviewer"
	Admin    UserRole = "Admin"
)

var BotCommands = []tg.BotCommand{
	{
		Command:     string(TakeVacation),
		Description: "Уйти в отпуск",
	},
	{
		Command:     string(ReturnToWork),
		Description: "Вернуться к работе",
	},
	{
		Command:     string(VacationsList),
		Description: "Показать список отпусков (только для админов)",
	},
	{
		Command:     string(SendToVacation),
		Description: "Отправить сотрудника в отпуск (только для админов)",
	},
}

const (
	DateFormatLayout = "02.01.2006"
)
