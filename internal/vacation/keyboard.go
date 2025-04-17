package vacation

import (
	"code-review-tg-bot/internal/access"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// createUsersKeyboard создает клавиатуру со списком пользователей
func (s *ServiceVacation) createUsersKeyboard(users access.UserIds) tg.ReplyKeyboardMarkup {
	var rows [][]tg.KeyboardButton

	for userId, userNameFromEnv := range users {
		// Пропускаем пользователей, которые уже в отпуске
		if s.IsUserOnVacation(userId) {
			continue
		}

		rows = append(rows, []tg.KeyboardButton{
			tg.NewKeyboardButton(fmt.Sprintf("👤 %s", userNameFromEnv)),
		})
	}

	// Добавляем кнопку отмены только если есть пользователи
	if len(rows) > 0 {
		rows = append(rows, []tg.KeyboardButton{
			tg.NewKeyboardButton(ButtonTextConstants.Cancel),
		})
	} else {
		// Если нет доступных пользователей, показываем только кнопку отмены
		rows = append(rows, []tg.KeyboardButton{
			tg.NewKeyboardButton(ButtonTextConstants.Cancel),
		})
	}

	keyboard := tg.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	keyboard.Selective = true

	return keyboard
}

// Создает клавиатуру с датами
func (s *ServiceVacation) createDateKeyboard(dates []string) tg.ReplyKeyboardMarkup {
	var rows [][]tg.KeyboardButton
	var elemCountInRow int = 3

	// Создаем ряды по 3 кнопки в каждом
	for i := 0; i < len(dates); i += elemCountInRow {
		var row []tg.KeyboardButton
		for j := 0; j < elemCountInRow && i+j < len(dates); j++ {
			row = append(row, tg.NewKeyboardButton(dates[i+j]))
		}
		rows = append(rows, row)
	}

	// Добавляем кнопку отмены в последний ряд
	rows = append(rows, []tg.KeyboardButton{
		tg.NewKeyboardButton(ButtonTextConstants.Cancel),
	})

	keyboard := tg.NewReplyKeyboard(rows...)
	keyboard.Selective = true
	return keyboard
}

// Создает клавиатуру с пользователями в отпуске
func (s *ServiceVacation) createVacationsListKeyboard() ([][]tg.KeyboardButton, error) {
	var buttons [][]tg.KeyboardButton

	allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге списка пользователей: %w", err)
	}

	// Проверяем статус отпуска для каждого пользователя
	for userId, userName := range allUsers {
		if s.IsUserOnVacation(userId) {

			// Добавляем две кнопки для каждого пользователя
			buttons = append(buttons,
				[]tg.KeyboardButton{
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", ButtonTextConstants.ReturnFromVacation, userName)),
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", ButtonTextConstants.ChangeVacation, userName)),
				},
			)
		}
	}

	if len(buttons) > 0 {
		buttons = append(buttons, []tg.KeyboardButton{tg.NewKeyboardButton(ButtonTextConstants.Cancel)})
	}

	return buttons, nil
}

// GetDefaultKeyboard возвращает клавиатуру по умолчанию
func (s *ServiceVacation) GetDefaultKeyboard(userId int64) tg.ReplyKeyboardMarkup {
	var defaultButtons []tg.KeyboardButton

	// Показываем кнопки отпуска только ревьюерам
	if access.HasVacationAccess(userId) {

		if s.IsUserOnVacation(userId) {
			defaultButtons = append(defaultButtons, tg.NewKeyboardButton(ButtonTextConstants.ReturnToWork))
		} else {
			defaultButtons = append(defaultButtons, tg.NewKeyboardButton(ButtonTextConstants.TakeVacation))
		}
	}

	// Показываем админ-кнопки только админам
	if access.HasAdminAccess(userId) {
		defaultButtons = append(defaultButtons,
			tg.NewKeyboardButton("📋 Список отпусков"),
			tg.NewKeyboardButton("➕ Отправить в отпуск"))
	}

	keyboard := tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(defaultButtons...),
		tg.NewKeyboardButtonRow(tg.NewKeyboardButton(ButtonTextConstants.Cancel)),
	)
	keyboard.Selective = true
	return keyboard
}
