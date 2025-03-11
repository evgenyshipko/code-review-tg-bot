package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/logger"
	"fmt"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обрабатывает текстовые команды, связанные с отпуском
func (s *ServiceVacation) handleTextCommand(update tg.Update, bot *tg.BotAPI) (executed bool) {
	if !access.HasVacationAccess(update.Message.From.ID) {
		return false
	}

	// Сбрасываем состояние админ-панели при работе с обычными командами
	s.resetAdminState(update.Message.From.ID)

	var commandMap = map[string]string{
		"отпуск": ButtonTakeVacation,
		"работа": ButtonReturnToWork,
	}
	text := strings.ToLower(update.Message.Text)

	for keyword, action := range commandMap {
		if strings.Contains(text, keyword) {
			// Устанавливаем соответствующее состояние
			if action == ButtonTakeVacation {

				if s.IsUserOnVacation(update.Message.From.ID) {
					returnDate := s.getVacationReturnDate(update.Message.From.ID)
					msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("@%s, Вы уже находитесь в отпуске до %s",
						update.Message.From.UserName,
						returnDate.Format(DateFormatLayout)))
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					return true
				}
				s.storage.Set(fmt.Sprintf("user_state_%d", update.Message.From.ID), UserStateSelectingDate)
			} else {
				s.ResetAllStates(update.Message.From.ID)
			}

			s.handleChangeVacationStatus(tg.Update{
				Message: &tg.Message{
					Text:      action,
					From:      update.Message.From,
					Chat:      update.Message.Chat,
					MessageID: update.Message.MessageID,
				},
			}, bot)
			return true
		}
	}

	return false
}

// Обрабатывает все команды, связанные с отпусками и админ-панелью
func (s *ServiceVacation) HandleCommand(update tg.Update, bot *tg.BotAPI) (executed bool) {
	if update.Message == nil && !update.Message.IsCommand() {
		return false
	}

	switch update.Message.Command() {
	case rest, work:
		if !access.HasVacationAccess(update.Message.From.ID) {
			msg := tg.NewMessage(update.Message.Chat.ID, "У вас нет доступа к этой команде")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return true
		}

		if update.Message.Command() == rest {
			// Сбрасываем состояние админ-панели
			s.resetAdminState(update.Message.From.ID)

			// Проверяем, не находится ли пользователь уже в отпуске
			if s.IsUserOnVacation(update.Message.From.ID) {
				returnDate := s.getVacationReturnDate(update.Message.From.ID)
				msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("@%s, Вы уже находитесь в отпуске до %s",
					update.Message.From.UserName,
					returnDate.Format(DateFormatLayout)))
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				return true
			}

			// Устанавливаем состояние выбора даты
			s.storage.Set(fmt.Sprintf("user_state_%d", update.Message.From.ID), UserStateSelectingDate)
			s.handleChangeVacationStatus(tg.Update{
				Message: &tg.Message{
					Text:      ButtonTakeVacation,
					From:      update.Message.From,
					Chat:      update.Message.Chat,
					MessageID: update.Message.MessageID,
				},
			}, bot)
			return true
		} else {
			// Сбрасываем все состояния
			s.ResetAllStates(update.Message.From.ID)
			s.handleChangeVacationStatus(tg.Update{
				Message: &tg.Message{
					Text:      ButtonReturnToWork,
					From:      update.Message.From,
					Chat:      update.Message.Chat,
					MessageID: update.Message.MessageID,
				},
			}, bot)
			return true
		}
	case vacations, vacations_start:
		if !access.HasAdminAccess(update.Message.From.ID) {
			msg := tg.NewMessage(update.Message.Chat.ID, "У вас нет доступа к этой команде")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return true
		}

		if update.Message.Command() == vacations {
			return s.handleVacationsCommand(update, bot)
		} else {
			return s.handleAdminCommand(update, bot)
		}
	}

	return false
}

// Обрабатывает нажатия на кнопки
func (s *ServiceVacation) handleButtonPress(update tg.Update, bot *tg.BotAPI) (executed bool) {
	if update.Message == nil {
		return false
	}

	var userState string
	s.storage.Get(fmt.Sprintf("user_state_%d", update.Message.From.ID), &userState)

	if strings.HasPrefix(update.Message.Text, ButtonReturnFromVacation) {
		userName := strings.TrimPrefix(update.Message.Text, ButtonReturnFromVacation)
		return s.handleReturnFromVacation(update, bot, userName)
	}

	if strings.HasPrefix(update.Message.Text, ButtonChangeVacation) {
		userName := strings.TrimPrefix(update.Message.Text, ButtonChangeVacation)
		return s.handleChangeVacation(update, bot, userName)
	}

	switch update.Message.Text {
	case ButtonVacationsList:
		if !access.IsAdmin(update.Message.From.ID) {
			return false
		}
		return s.handleVacationsCommand(update, bot)
	case ButtonVacationsStart:
		if !access.IsAdmin(update.Message.From.ID) {
			return false
		}
		return s.handleAdminCommand(update, bot)
	case ButtonTakeVacation, ButtonChangeVacation:
		// Сбрасываем состояние админ-панели
		s.resetAdminState(update.Message.From.ID)
		s.handleChangeVacationStatus(update, bot)
		return true
	case ButtonReturnToWork, ButtonCancel:
		// Сбрасываем все состояния
		s.ResetAllStates(update.Message.From.ID)
		s.handleChangeVacationStatus(update, bot)
		return true
	default:
		// Проверяем, является ли сообщение датой после выбора даты в клавиатуре
		if userState == UserStateSelectingDate && s.isDateFormat(update.Message.Text) {
			// Сбрасываем состояние админ-панели
			s.resetAdminState(update.Message.From.ID)
			s.handleChangeVacationStatus(update, bot)
			return true
		}
	}

	return false
}

// Обрабатывает изменение статуса отпуска
func (s *ServiceVacation) handleChangeVacationStatus(update tg.Update, bot *tg.BotAPI) {
	switch update.Message.Text {
	case ButtonTakeVacation, ButtonChangeVacation:
		// Показываем календарь на клавиатуре
		dates := s.generateVacationDates()
		keyboard := s.createDateKeyboard(dates)

		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите дату выхода на работу:")
		msg.ReplyMarkup = keyboard
		msg.ReplyToMessageID = update.Message.MessageID

		// Устанавливаем состояние выбора даты
		s.storage.Set(fmt.Sprintf("user_state_%d", update.Message.From.ID), UserStateSelectingDate)

		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки календаря", "error", err)
		}

	case ButtonReturnToWork:
		// Проверяем, находится ли пользователь в отпуске
		if !s.IsUserOnVacation(update.Message.From.ID) {
			msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("@%s, Вы не находитесь в отпуске", update.Message.From.UserName))
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
			_, err := bot.Send(msg)
			if err != nil {
				logger.Instance.Error("Ошибка отправки сообщения", "error", err)
			}
			return
		}

		s.endVacation(update.Message.From.ID)

		// Сбрасываем состояние пользователя
		s.storage.Set(fmt.Sprintf("user_state_%d", update.Message.From.ID), UserStateNone)

		message := fmt.Sprintf("@%s вернулся к работе", update.Message.From.UserName)
		msg := tg.NewMessage(update.Message.Chat.ID, message)
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		msg.ReplyToMessageID = update.Message.MessageID
		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки статуса", "error", err)
		}

	case ButtonCancel:
		// Сбрасываем состояние пользователя
		s.storage.Set(fmt.Sprintf("user_state_%d", update.Message.From.ID), UserStateNone)

		msg := tg.NewMessage(update.Message.Chat.ID, MsgKeyboardClosed)
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		msg.ReplyToMessageID = update.Message.MessageID
		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки клавиатуры", "error", err)
		}

	default:
		// Проверяем состояние пользователя
		var userState string
		s.storage.Get(fmt.Sprintf("user_state_%d", update.Message.From.ID), &userState)

		// Проверяем, является ли сообщение датой и находится ли пользователь в состоянии выбора даты
		if userState == UserStateSelectingDate {
			returnDate, err := time.Parse(DateFormatLayout, update.Message.Text)
			if err == nil {
				if !s.isValidVacationDate(returnDate) {
					msg := tg.NewMessage(update.Message.Chat.ID, "Дата выхода на работу должна быть не раньше завтрашнего дня и не позже чем через 3 недели")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					return
				}

				// Обновляем статус в бд
				if err := s.startVacation(update.Message.From.ID, returnDate); err != nil {
					logger.Instance.Error("Ошибка сохранения статуса отпуска", "error", err)
					return
				}

				// Сбрасываем состояние пользователя
				s.storage.Set(fmt.Sprintf("user_state_%d", update.Message.From.ID), UserStateNone)

				message := fmt.Sprintf("@%s ушёл в отпуск до %s", update.Message.From.UserName, update.Message.Text)
				msg := tg.NewMessage(update.Message.Chat.ID, message)
				msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
				msg.ReplyToMessageID = update.Message.MessageID
				_, err := bot.Send(msg)

				if err != nil {
					logger.Instance.Error("Ошибка отправки статуса", "error", err)
				}
			}
		}
	}
}
