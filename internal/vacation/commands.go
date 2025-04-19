package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/logger"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

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
	case constants.TestTakeVacation:
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
	case constants.TakeVacation:
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
	case constants.ReturnToWork:
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
	case constants.VacationsList:
		return s.handleVacationListCommand(update, bot)
	case constants.SendOnVacation:
		return s.handleSendOnVacation(update, bot)
	}

	return false
}
