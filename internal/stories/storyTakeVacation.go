package stories

import (
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/dates"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/vacation"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

var ShowDatesKeyboardStageCb = func(update tg.Update, s *StoryService) ExecutionStatus {

	userId := update.Message.From.ID

	if s.vacationService.IsUserOnVacation(userId) {
		returnDate := s.vacationService.GetVacationReturnDate(userId)
		msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("@%s, Вы уже находитесь в отпуске до %s",
			update.Message.From.UserName,
			returnDate.Format(constants.DateFormatLayout)))
		msg.ReplyToMessageID = update.Message.MessageID
		s.bot.Send(msg)
		return Dropped
	}

	dates := dates.GenerateVacationDates()
	keyboard := vacation.CreateDateKeyboard(dates)

	msg := tg.NewMessage(update.Message.Chat.ID, "Выберите дату выхода на работу:")
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := s.bot.Send(msg)
	if err != nil {
		logger.Instance.Error("Ошибка отправки календаря", "error", err)
	}
	return Executed
}

var PickKeyboardDateCb = func(update tg.Update, s *StoryService) ExecutionStatus {

	var message *tg.Message = update.Message
	var text string = message.Text

	if text == constants.ButtonTextConstants.Cancel {
		msg := tg.NewMessage(message.Chat.ID, "завершение процесса выбора даты")
		msg.ReplyToMessageID = message.MessageID
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		s.bot.Send(msg)
		return Dropped
	}

	logger.Instance.Infow("PickKeyboardDateCb", "message", message)

	//TODO: вынести валидацию даты в отдельную функцию
	returnDate, err := time.Parse(constants.DateFormatLayout, text)

	if err != nil {
		msg := tg.NewMessage(message.Chat.ID, "Необходимо выбрать дату на клавиатуре")
		msg.ReplyToMessageID = message.MessageID
		s.bot.Send(msg)
		return NotExecuted
	}

	if !dates.IsValidVacationDate(returnDate) {
		msg := tg.NewMessage(message.Chat.ID, "Дата выхода на работу должна быть не раньше завтрашнего дня и не позже чем через 3 недели")
		msg.ReplyToMessageID = message.MessageID
		s.bot.Send(msg)
		return NotExecuted
	}

	if err := s.vacationService.StartVacation(message.From.ID, returnDate); err != nil {
		logger.Instance.Error("Ошибка сохранения статуса отпуска", "error", err)
		msg := tg.NewMessage(message.Chat.ID, "Ошибка сохранения статуса отпуска")
		msg.ReplyToMessageID = message.MessageID
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		s.bot.Send(msg)
		return Dropped
	}

	str := fmt.Sprintf("@%s ушёл в отпуск до %s", message.From.UserName, text)
	msg := tg.NewMessage(message.Chat.ID, str)
	msg.ReplyToMessageID = message.MessageID
	msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
	_, err = s.bot.Send(msg)
	if err != nil {
		logger.Instance.Error("Ошибка отправки статуса", "error", err)
	}
	return Executed
}

var ShowDatesKeyboardStage = NewStage("ShowDatesKeyboard", ShowDatesKeyboardStageCb)
var PickKeyboardDateStage = NewStage("PickKeyboardDate", PickKeyboardDateCb)

var TakeVacationStory = NewStory("TakeVacationStory", []Stage{*ShowDatesKeyboardStage, *PickKeyboardDateStage})
