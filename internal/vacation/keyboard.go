package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *VacationService) createUsersKeyboard(users access.UserIds) tg.ReplyKeyboardMarkup {
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

	rows = append(rows, []tg.KeyboardButton{
		tg.NewKeyboardButton(constants.ButtonTextConstants.Cancel),
	})

	keyboard := tg.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	keyboard.Selective = true

	return keyboard
}

func (s *VacationService) createDateKeyboard(dates []string) tg.ReplyKeyboardMarkup {
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

	rows = append(rows, []tg.KeyboardButton{
		tg.NewKeyboardButton(constants.ButtonTextConstants.Cancel),
	})

	keyboard := tg.NewReplyKeyboard(rows...)
	keyboard.Selective = true
	return keyboard
}

func (s *VacationService) createVacationsListKeyboard() ([][]tg.KeyboardButton, error) {
	var buttons [][]tg.KeyboardButton

	allUsers := s.usersMap.ReviewersIdsMap

	// Проверяем статус отпуска для каждого пользователя
	for userId, userName := range allUsers {
		if s.IsUserOnVacation(userId) {

			// Добавляем две кнопки для каждого пользователя
			buttons = append(buttons,
				[]tg.KeyboardButton{
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", constants.ButtonTextConstants.ReturnFromVacation, userName)),
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", constants.ButtonTextConstants.ChangeVacation, userName)),
				},
			)
		}
	}

	if len(buttons) > 0 {
		buttons = append(buttons, []tg.KeyboardButton{tg.NewKeyboardButton(constants.ButtonTextConstants.Cancel)})
	}

	return buttons, nil
}

func (s *VacationService) getDefaultKeyboard(userId int64) tg.ReplyKeyboardMarkup {
	var defaultButtons []tg.KeyboardButton

	// Показываем кнопки отпуска только ревьюерам
	if access.HasVacationAccess(userId, *s.usersMap) {

		if s.IsUserOnVacation(userId) {
			defaultButtons = append(defaultButtons, tg.NewKeyboardButton(constants.ButtonTextConstants.ReturnToWork))
		} else {
			defaultButtons = append(defaultButtons, tg.NewKeyboardButton(constants.ButtonTextConstants.TakeVacation))
		}
	}

	// Показываем админ-кнопки только админам
	if access.HasAdminAccess(userId, *s.usersMap) {
		defaultButtons = append(defaultButtons,
			tg.NewKeyboardButton("📋 Список отпусков"),
			tg.NewKeyboardButton("➕ Отправить в отпуск"))
	}

	keyboard := tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(defaultButtons...),
		tg.NewKeyboardButtonRow(tg.NewKeyboardButton(constants.ButtonTextConstants.Cancel)),
	)
	keyboard.Selective = true
	return keyboard
}
