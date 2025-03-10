package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/logger"
	"fmt"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// createUsersKeyboard создает клавиатуру со списком пользователей
func (s *ServiceVacation) createUsersKeyboard(users map[string]int64, chatID int64) tg.ReplyKeyboardMarkup {
	var rows [][]tg.KeyboardButton

	for _, userId := range users {
		// Проверяем, не находится ли пользователь в отпуске
		var status StatusVacationUser
		exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &status)

		// Пропускаем пользователей, которые уже в отпуске
		if exists && status.IsOnVacation {
			continue
		}

		member, err := s.bot.GetChatMember(tg.GetChatMemberConfig{
			ChatConfigWithUser: tg.ChatConfigWithUser{
				ChatID: chatID,
				UserID: userId,
			},
		})

		if err != nil {
			logger.Instance.Error("Ошибка при получении информации о пользователе", "error", err)
			continue
		}

		fullName := strings.TrimSpace(member.User.FirstName + " " + member.User.LastName)
		rows = append(rows, []tg.KeyboardButton{
			tg.NewKeyboardButton(fmt.Sprintf("👤 %s", fullName)),
		})
	}

	// Добавляем кнопку отмены только если есть пользователи
	if len(rows) > 0 {
		rows = append(rows, []tg.KeyboardButton{
			tg.NewKeyboardButton(ButtonCancel),
		})
	} else {
		// Если нет доступных пользователей, показываем только кнопку отмены
		rows = append(rows, []tg.KeyboardButton{
			tg.NewKeyboardButton(ButtonCancel),
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
		tg.NewKeyboardButton(ButtonCancel),
	})

	keyboard := tg.NewReplyKeyboard(rows...)
	keyboard.Selective = true
	return keyboard
}

// Создает клавиатуру с действиями для выбранного пользователя
func (s *ServiceVacation) createAdminActionsKeyboard(userId int64) tg.ReplyKeyboardMarkup {
	var buttons [][]tg.KeyboardButton
	var actionButtons []tg.KeyboardButton

	// Проверяем статус отпуска пользователя
	var status StatusVacationUser
	exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &status)

	if exists && status.IsOnVacation {
		// Если пользователь в отпуске, показываем только кнопку возврата
		actionButtons = append(actionButtons, tg.NewKeyboardButton(ButtonChangeVacation), tg.NewKeyboardButton(ButtonReturnToWork))
	} else {
		// Если не в отпуске, показываем только кнопку добавления в отпуск
		actionButtons = append(actionButtons, tg.NewKeyboardButton(ButtonTakeVacation))
	}

	buttons = append(buttons, actionButtons)
	buttons = append(buttons, []tg.KeyboardButton{tg.NewKeyboardButton(ButtonCancel)})

	keyboard := tg.NewReplyKeyboard(buttons...)
	keyboard.OneTimeKeyboard = true
	keyboard.Selective = true

	return keyboard
}

// Создает клавиатуру с пользователями в отпуске
func (s *ServiceVacation) createVacationsListKeyboard(update tg.Update) ([][]tg.KeyboardButton, error) {
	var buttons [][]tg.KeyboardButton

	allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге списка пользователей: %w", err)
	}

	// Проверяем статус отпуска для каждого пользователя
	for _, userId := range allUsers {
		var status StatusVacationUser
		exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &status)

		if exists && status.IsOnVacation {
			member, err := s.bot.GetChatMember(tg.GetChatMemberConfig{
				ChatConfigWithUser: tg.ChatConfigWithUser{
					ChatID: update.Message.Chat.ID,
					UserID: userId,
				},
			})

			if err != nil {
				continue
			}

			// Добавляем две кнопки для каждого пользователя
			fullName := fmt.Sprintf("%s %s", member.User.FirstName, member.User.LastName)
			buttons = append(buttons,
				[]tg.KeyboardButton{
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", ButtonReturnFromVacation, fullName)),
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", ButtonChangeVacation, fullName)),
				},
			)
		}
	}

	if len(buttons) > 0 {
		buttons = append(buttons, []tg.KeyboardButton{tg.NewKeyboardButton(ButtonCancel)})
	}

	return buttons, nil
}

// GetDefaultKeyboard возвращает клавиатуру по умолчанию
func (s *ServiceVacation) GetDefaultKeyboard(userId int64) tg.ReplyKeyboardMarkup {
	var defaultButtons []tg.KeyboardButton

	// Показываем кнопки отпуска только ревьюерам
	if access.HasVacationAccess(userId) {
		var status StatusVacationUser
		exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &status)

		if exists && status.IsOnVacation {
			defaultButtons = append(defaultButtons, tg.NewKeyboardButton(ButtonReturnToWork))
		} else {
			defaultButtons = append(defaultButtons, tg.NewKeyboardButton(ButtonTakeVacation))
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
		tg.NewKeyboardButtonRow(tg.NewKeyboardButton(ButtonCancel)),
	)
	keyboard.Selective = true
	return keyboard
}
