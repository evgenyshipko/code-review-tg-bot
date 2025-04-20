package stories

import (
	"code-review-tg-bot/internal/logger"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var ReturnToWorkCb = func(update tg.Update, s *StoryService) ExecutionStatus {
	if !s.vacationService.IsUserOnVacation(update.Message.From.ID) {
		msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("@%s, Вы не находитесь в отпуске", update.Message.From.UserName))
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		_, err := s.bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки сообщения", "error", err)
		}
		return Dropped
	}

	s.vacationService.EndVacation(update.Message.From.ID)

	message := fmt.Sprintf("@%s вернулся к работе", update.Message.From.UserName)
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := s.bot.Send(msg)
	if err != nil {
		logger.Instance.Error("Ошибка отправки статуса", "error", err)
	}
	return Executed
}

var ReturnToWorkStage = NewStage("ReturnToWorkStage", ReturnToWorkCb)

var ReturnToWorkStory = NewStory("ReturnToWorkStory", []Stage{*ReturnToWorkStage})
