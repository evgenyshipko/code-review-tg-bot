package vacation

// Константы для текста кнопок
const (
	ButtonTakeVacation       = "🏃 Оформить отпуск"
	ButtonReturnFromVacation = "🔄 Вернуть из отпуска:"
	ButtonChangeVacation     = "📅 Изменить отпуск:"
	ButtonReturnToWork       = "💼 Выход на работу"
	ButtonCancel             = "❌ Закрыть"
	ButtonVacationsList      = "📋 Список отпусков"
	ButtonVacationsLeave     = "➕ Отправить в отпуск"
)

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
