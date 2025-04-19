package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/logger"
	"fmt"
	"reflect"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обрабатывает текстовые команды, связанные с отпуском
func (s *VacationService) handleTextCommand(update tg.Update, bot *tg.BotAPI) (executed bool) {
	if !access.HasVacationAccess(update.Message.From.ID, *s.usersMap) {
		return false
	}

	msg := strings.ToLower(update.Message.Text)
	wordsFromMsg := strings.Fields(msg)

	// Строгая проверка на команду
	if len(wordsFromMsg) != 2 || wordsFromMsg[0] != "@"+strings.ToLower(bot.Self.UserName) {
		return false
	}

	var commandMap = map[string]string{
		"отпуск": constants.ButtonTextConstants.TakeVacation,
		"работа": constants.ButtonTextConstants.ReturnToWork,
	}

	if action, ok := commandMap[wordsFromMsg[1]]; ok {
		// Сбрасывание состояние админ-панели при работе с обычными командами
		s.resetAdminState(update.Message.From.ID)

		// Устанавливаем соответствующее состояние
		if action == constants.ButtonTextConstants.TakeVacation {
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

	return false
}

func (s *VacationService) HandleCommand(update tg.Update, bot *tg.BotAPI) (executed bool) {
	command := update.Message.Command()
	if !access.IsUserHasAccessToCommand(update.Message.From.ID, command, *s.users) {
		msg := tg.NewMessage(update.Message.Chat.ID, "У Вас нет доступа к данной команде")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return true
	}

	switch command {
	case constants.Start:
		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите команду:")
		msg.ReplyMarkup = s.getDefaultKeyboard(update.Message.From.ID)
		msg.ReplyToMessageID = update.Message.MessageID

		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки клавиатуры", "error", err)
		}
		return true
	case constants.Test_vacation:
		if s.IsUserOnVacation(update.Message.From.ID) {
			returnDate := s.getVacationReturnDate(update.Message.From.ID)
			msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("@%s, Вы уже находитесь в отпуске до %s",
				update.Message.From.UserName,
				returnDate.Format(DateFormatLayout)))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return true
		}

		// Устанавливаем отпуск на 1 минуту
		returnDate := time.Now().Add(time.Minute)
		err := s.startVacation(update.Message.From.ID, returnDate)
		if err != nil {
			msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка при установке отпуска: %s", err))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return true
		}

		message := fmt.Sprintf("@%s ушел в тестовый отпуск на 1 минуту (до %s)",
			update.Message.From.UserName,
			returnDate.Format(DateFormatLayout))
		msg := tg.NewMessage(update.Message.Chat.ID, message)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return true
	case constants.Rest:
		{
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
					Text:      constants.ButtonTextConstants.TakeVacation,
					From:      update.Message.From,
					Chat:      update.Message.Chat,
					MessageID: update.Message.MessageID,
				},
			}, bot)
			return true
		}
	case constants.Work:
		{
			// Сбрасываем все состояния
			s.ResetAllStates(update.Message.From.ID)
			s.handleChangeVacationStatus(tg.Update{
				Message: &tg.Message{
					Text:      constants.ButtonTextConstants.ReturnToWork,
					From:      update.Message.From,
					Chat:      update.Message.Chat,
					MessageID: update.Message.MessageID,
				},
			}, bot)
			return true
		}
	case constants.Vacations:
		return s.handleVacationsCommand(update, bot)
	case constants.Vacations_start:
		return s.handleAdminCommand(update, bot)
	}

	return false
}

func (s *VacationService) isButtonPress(text string) bool {

	val := reflect.ValueOf(constants.ButtonTextConstants)

	for i := 0; i < val.NumField(); i++ {
		if strings.Contains(text, val.Field(i).String()) {
			return true
		}
	}

	return false
}

// Обрабатывает нажатия на кнопки
func (s *VacationService) handleButtonPress(update tg.Update, bot *tg.BotAPI) (executed bool) {

	var userState string
	s.storage.Get(fmt.Sprintf("user_state_%d", update.Message.From.ID), &userState)

	if strings.HasPrefix(update.Message.Text, constants.ButtonTextConstants.ReturnFromVacation) {
		userName := strings.TrimPrefix(update.Message.Text, constants.ButtonTextConstants.ReturnFromVacation)
		return s.handleReturnFromVacation(update, bot, userName)
	}

	if strings.HasPrefix(update.Message.Text, constants.ButtonTextConstants.ChangeVacation) {
		userName := strings.TrimPrefix(update.Message.Text, constants.ButtonTextConstants.ChangeVacation)
		return s.handleChangeVacation(update, bot, userName)
	}

	switch update.Message.Text {
	case constants.ButtonTextConstants.VacationsList:
		if !access.IsAdmin(update.Message.From.ID, *s.usersMap) {
			return false
		}
		return s.handleVacationsCommand(update, bot)
	case constants.ButtonTextConstants.VacationsStart:
		if !access.IsAdmin(update.Message.From.ID, *s.usersMap) {
			return false
		}
		return s.handleAdminCommand(update, bot)
	case constants.ButtonTextConstants.TakeVacation, constants.ButtonTextConstants.ChangeVacation:
		// Сбрасываем состояние админ-панели
		s.resetAdminState(update.Message.From.ID)
		s.handleChangeVacationStatus(update, bot)
		return true
	case constants.ButtonTextConstants.ReturnToWork, constants.ButtonTextConstants.Cancel:
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
func (s *VacationService) handleChangeVacationStatus(update tg.Update, bot *tg.BotAPI) {
	switch update.Message.Text {
	case constants.ButtonTextConstants.TakeVacation, constants.ButtonTextConstants.ChangeVacation:
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

	case constants.ButtonTextConstants.ReturnToWork:
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

	case constants.ButtonTextConstants.Cancel:
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
