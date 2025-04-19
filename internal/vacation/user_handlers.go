package vacation

import (
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/logger"
	"fmt"
	"reflect"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *VacationService) isButtonPress(text string) bool {

	val := reflect.ValueOf(constants.ButtonTextConstants)

	for i := 0; i < val.NumField(); i++ {
		if strings.Contains(text, val.Field(i).String()) {
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
