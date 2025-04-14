package vacation

type ButtonText struct {
	TakeVacation       string
	ReturnFromVacation string
	ChangeVacation     string
	ReturnToWork       string
	Cancel             string
	VacationsList      string
	VacationsStart     string
}

// Константы для текста кнопок
var ButtonTextConstants = ButtonText{
	TakeVacation:       "🏃 Оформить отпуск",
	ReturnFromVacation: "🔄 Вернуть из отпуска:",
	ChangeVacation:     "📅 Изменить отпуск:",
	ReturnToWork:       "💼 Выход на работу",
	Cancel:             "❌ Закрыть",
	VacationsList:      "📋 Список отпусков",
	VacationsStart:     "➕ Отправить в отпуск",
}

func (b ButtonText) GetHashMap() map[string]bool {
	return map[string]bool{
		b.TakeVacation:       true,
		b.ReturnFromVacation: true,
		b.ChangeVacation:     true,
		b.ReturnToWork:       true,
		b.Cancel:             true,
		b.VacationsList:      true,
		b.VacationsStart:     true,
	}
}

// Константы состояний админ-панели
const (
	AdminStateNone        = ""             // начальное состояние
	AdminStateUserList    = "user_list"    // список пользователей
	AdminStateUserActions = "user_actions" // действия с пользователем
	AdminStateSetVacation = "set_vacation" // установка отпуска
)

// Константы состояний пользователя
const (
	UserStateNone          = ""            // начальное состояние
	UserStateSelectingDate = "select_date" // выбор даты отпуска
)

// Константы для форматирования даты
const (
	DateFormatLayout = "02.01.2006"
)

// Константы для сообщений
const (
	MsgKeyboardClosed = "Клавиатура закрыта"
)

// Константы для комманд
const (
	rest            = "rest"
	work            = "work"
	vacations_start = "vacations_start"
	vacations       = "vacations"
	test_vacation   = "test_vacation"
)
