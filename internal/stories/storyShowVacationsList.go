package stories

import (
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/dates"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/vacation"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
	"time"
)

var ShowVacationsListAndKeyboardCb = func(update tg.Update, s *StoryService) ExecutionStatus {

	users := s.vacationService.GetUsersInVacation()
	if len(users) == 0 {
		msg := tg.NewMessage(update.Message.Chat.ID, "В данный момент нет пользователей в отпуске")
		s.bot.Send(msg)
		return Finished
	}

	message := s.vacationService.GetUsersInVacationMessage(users)

	buttons, err := s.vacationService.CreateVacationsListKeyboard()
	if err != nil {
		logger.Instance.Error("Ошибка при создании клавиатуры", "error", err)
		return Dropped
	}

	msg := tg.NewMessage(update.Message.Chat.ID, message)

	if len(buttons) > 0 {
		keyboard := tg.NewReplyKeyboard(buttons...)
		msg.ReplyMarkup = keyboard
		msg.ReplyToMessageID = update.Message.MessageID
	}

	s.bot.Send(msg)
	return Executed
}

var PushVacationKeyboardStageCb = func(update tg.Update, s *StoryService) ExecutionStatus {
	if update.Message.Text == constants.ButtonTextConstants.Cancel {
		msg := tg.NewMessage(update.Message.Chat.ID, "Клавиатура закрыта")
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		s.bot.Send(msg)
		return Dropped
	}

	if strings.HasPrefix(update.Message.Text, constants.ButtonTextConstants.ReturnFromVacation) {
		userName := strings.TrimPrefix(update.Message.Text, constants.ButtonTextConstants.ReturnFromVacation)

		allUsers := s.vacationService.UsersMap.ReviewersIdsMap
		foundUserId := s.vacationService.FindVacationUserIdByName(allUsers, userName)

		if foundUserId == 0 {
			logger.Instance.Warnf("HandleVacationListPushBtnCb пользователь %s не найден", userName)
			return Dropped
		}

		s.vacationService.SetUserReturnedFromVacation(foundUserId, userName, update, s.bot)
		return Finished
	}

	if strings.HasPrefix(update.Message.Text, constants.ButtonTextConstants.ChangeVacation) {
		userName := strings.TrimPrefix(update.Message.Text, constants.ButtonTextConstants.ChangeVacation)

		allUsers := s.vacationService.UsersMap.ReviewersIdsMap

		foundUserId := s.vacationService.FindVacationUserIdByName(allUsers, userName)
		if foundUserId == 0 {
			logger.Instance.Warnf("HandleVacationListPushBtnCb пользователь %s не найден", userName)
			return Dropped
		}

		s.setCurrentStoryExtraData(update.Message.From.ID, foundUserId)

		dates := dates.GenerateVacationDates()
		keyboard := vacation.CreateDateKeyboard(dates)

		msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Выберите новую дату выхода на работу для %s:", userName))
		msg.ReplyMarkup = keyboard
		msg.ReplyToMessageID = update.Message.MessageID

		s.bot.Send(msg)
		return Executed
	}

	msg := tg.NewMessage(update.Message.Chat.ID, "Выберите действие или закройте клавиатуру")
	msg.ReplyToMessageID = update.Message.MessageID
	s.bot.Send(msg)

	return NotExecuted
}

var AdminSetVacationCb = func(update tg.Update, s *StoryService) ExecutionStatus {
	returnDate, err := time.Parse(constants.DateFormatLayout, update.Message.Text)

	if err != nil {
		msg := tg.NewMessage(update.Message.Chat.ID, "Дата выхода на работу должна быть в формате ДД.ММ.ГГГГ")
		msg.ReplyToMessageID = update.Message.MessageID
		s.bot.Send(msg)
		return NotExecuted
	}

	if !dates.IsValidVacationDate(returnDate) {
		msg := tg.NewMessage(update.Message.Chat.ID, "Дата выхода на работу должна быть не раньше завтрашнего дня и не позже чем через 3 недели")
		msg.ReplyToMessageID = update.Message.MessageID
		s.bot.Send(msg)
		return NotExecuted
	}

	allUsers := s.vacationService.UsersMap.ReviewersIdsMap

	pickedUserId, err := getPickedUserIdFromExtraData(s, update.Message.From.ID)
	if err != nil {
		logger.Instance.Warn("getCurrentStoryExtraData", "err", err)
		return Dropped
	}

	var username string
	for uid, userNameFromEnv := range allUsers {
		if uid == pickedUserId {
			username = userNameFromEnv
			break
		}
	}

	if username == "" {
		logger.Instance.Warnw("Не удалось найти пользователя", "user_id", pickedUserId)
		return Dropped
	}

	if err := s.vacationService.StartVacation(pickedUserId, returnDate); err != nil {
		logger.Instance.Warnw("Ошибка сохранения статуса отпуска", "error", err)
		return Dropped
	}

	message := fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> добавлен в отпуск до %s",
		pickedUserId,
		username,
		update.Message.Text)
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
	msg.ReplyToMessageID = update.Message.MessageID

	s.bot.Send(msg)
	return Executed
}

var ShowVacationsListAndKeyboardStage = NewStage("ShowVacationsListAndKeyboardStage", ShowVacationsListAndKeyboardCb)

var PushVacationKeyboardStage = NewStage("PushVacationKeyboardStage", PushVacationKeyboardStageCb)

var AdminSetVacationStage = NewStage("AdminSetVacationStage", AdminSetVacationCb)

var ShowVacationListStory = NewStory("ShowVacationListStory", []Stage{*ShowVacationsListAndKeyboardStage, *PushVacationKeyboardStage, *AdminSetVacationStage})
