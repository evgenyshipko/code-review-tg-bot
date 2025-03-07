package vacation

// Константы для текста кнопок
const (
	ButtonTakeVacation   = "🏃 Оформить отпуск"
	ButtonReturnToWork   = "💼 Выход на работу"
	ButtonCancel         = "❌ Закрыть"
	ButtonChangeVacation = "📅 Изменить отпуск"
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
