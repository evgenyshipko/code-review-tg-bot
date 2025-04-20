package stories

import (
	"code-review-tg-bot/internal/dates"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/vacation"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

var getUsersKeyboardCb = func(update tg.Update, s *StoryService) ExecutionStatus {
	allUsers := s.vacationService.UsersMap.ReviewersIdsMap

	keyboard := s.vacationService.CreateUsersKeyboard(allUsers)
	msg := tg.NewMessage(update.Message.Chat.ID, "Выберите пользователя:")
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := s.bot.Send(msg)
	if err != nil {
		logger.Instance.Errorw("Ошибка отправки клавиатуры", "error", err)
		return Dropped
	}

	return Executed
}

var handlePickUserOnKeyboardCb = func(update tg.Update, s *StoryService) ExecutionStatus {
	prefix := "👤 "

	if !strings.HasPrefix(update.Message.Text, prefix) {
		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите пользователя на клавиатуре")
		msg.ReplyToMessageID = update.Message.MessageID
		s.bot.Send(msg)

		return NotExecuted
	}

	selectedFullName := strings.TrimPrefix(update.Message.Text, prefix)

	//TODO: перенести userMaps в storyService
	allUsers := s.vacationService.UsersMap.ReviewersIdsMap

	var selectedUserId int64
	for uid, userNameFromEnv := range allUsers {
		if userNameFromEnv == selectedFullName {
			selectedUserId = uid
			break
		}
	}

	if selectedUserId == 0 {
		msg := tg.NewMessage(update.Message.Chat.ID, "Пользователь не найден")
		msg.ReplyToMessageID = update.Message.MessageID
		s.bot.Send(msg)
		return Dropped
	}

	dates := dates.GenerateVacationDates()
	keyboard := vacation.CreateDateKeyboard(dates)

	msg := tg.NewMessage(update.Message.Chat.ID, "Выберите дату выхода на работу:")
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = update.Message.MessageID

	s.setCurrentStoryExtraData(update.Message.From.ID, selectedUserId)

	s.bot.Send(msg)
	return Executed
}

var getUsersKeyboardStage = NewStage("getUsersKeyboardStage", getUsersKeyboardCb)

var handlePickUserOnKeyboardStage = NewStage("handlePickUserOnKeyboardStage", handlePickUserOnKeyboardCb)

var SendToVacationStory = NewStory("SendToVacationStory", []Stage{*getUsersKeyboardStage, *handlePickUserOnKeyboardStage, *AdminSetVacationStage})
