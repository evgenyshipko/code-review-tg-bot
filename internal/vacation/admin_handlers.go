package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/logger"
	"fmt"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обрабатывает действия в админ-панели
func (s *ServiceVacation) HandleAdminPanel(update tg.Update, bot *tg.BotAPI, state AdminPanelState) {
	// Сбрасываем состояние пользователя при работе с админ-панелью
	s.ResetAllStates(update.Message.From.ID)

	if update.Message.Text == ButtonTextConstants.Cancel {
		// Сбрасываем состояние админ-панели
		s.resetAdminState(update.Message.From.ID)

		msg := tg.NewMessage(update.Message.Chat.ID, MsgKeyboardClosed)
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
		return
	}

	switch state.State {
	case AdminStateUserList:
		s.handleAdminPickUser(update, bot, state)
	case AdminStateUserActions:
		s.handleAdminActionsForUser(update, bot, state)
	case AdminStateSetVacation:
		s.handleAdminSetVacation(update, bot, state)
	}
}

// Обрабатывает выбор пользователя
func (s *ServiceVacation) handleAdminPickUser(update tg.Update, bot *tg.BotAPI, state AdminPanelState) {
	if strings.HasPrefix(update.Message.Text, "👤 ") {
		selectedFullName := strings.TrimPrefix(update.Message.Text, "👤 ")

		// Получаем ID пользователя по полному имени
		allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
		if err != nil {
			logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
			return
		}

		// Получаем ID выбранного пользователя
		var selectedUserId int64
		for userNameFromEnv, uid := range allUsers {

			if userNameFromEnv == selectedFullName {
				selectedUserId = uid
				break
			}
		}

		if selectedUserId == 0 {
			msg := tg.NewMessage(update.Message.Chat.ID, "Пользователь не найден")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return
		}

		// Показываем календарь для выбора даты
		dates := s.generateVacationDates()
		keyboard := s.createDateKeyboard(dates)

		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите дату выхода на работу:")
		msg.ReplyMarkup = keyboard
		msg.ReplyToMessageID = update.Message.MessageID

		// Обновляем состояние
		state.State = AdminStateSetVacation
		state.SelectedUID = selectedUserId
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), state)

		bot.Send(msg)
	}
}

// Обрабатывает действия в админ-панели
func (s *ServiceVacation) handleAdminActionsForUser(update tg.Update, bot *tg.BotAPI, state AdminPanelState) {
	switch update.Message.Text {
	case ButtonTextConstants.TakeVacation, ButtonTextConstants.ChangeVacation:
		// Показываем календарь
		dates := s.generateVacationDates()
		keyboard := s.createDateKeyboard(dates)

		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите дату выхода на работу:")
		msg.ReplyMarkup = keyboard
		msg.ReplyToMessageID = update.Message.MessageID

		// Обновляем состояние
		state.State = AdminStateSetVacation
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), state)

		bot.Send(msg)

	case ButtonTextConstants.ReturnToWork:
		// Проверяем, находится ли пользователь в отпуске
		if !s.IsUserOnVacation(update.Message.From.ID) {
			msg := tg.NewMessage(update.Message.Chat.ID, "Пользователь не находится в отпуске")
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
			bot.Send(msg)
			return
		}

		// Получаем список пользователей для поиска имени
		allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
		if err != nil {
			logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
			return
		}

		// Ищем имя пользователя по ID
		var username string
		for userNameFromEnv, uid := range allUsers {
			if uid == state.SelectedUID {

				username = userNameFromEnv
				break
			}
		}

		if username == "" {
			logger.Instance.Error("Не удалось найти имя пользователя", "user_id", state.SelectedUID)
			return
		}

		s.endVacation(state.SelectedUID)

		message := fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> возвращен на работу",
			state.SelectedUID,
			username)
		msg := tg.NewMessage(update.Message.Chat.ID, message)
		msg.ParseMode = tg.ModeHTML
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		msg.ReplyToMessageID = update.Message.MessageID

		// Сбрасываем состояние админ-панели
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), AdminPanelState{State: AdminStateNone})

		bot.Send(msg)
	}
}

// Обрабатывает добавление пользователя в отпуск
func (s *ServiceVacation) handleAdminSetVacation(update tg.Update, bot *tg.BotAPI, state AdminPanelState) {
	// Проверка, является ли сообщение датой
	returnDate, err := time.Parse(DateFormatLayout, update.Message.Text)

	if err != nil {
		msg := tg.NewMessage(update.Message.Chat.ID, "Дата выхода на работу должна быть в формате ДД.ММ.ГГГГ")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	// Проверяем валидность даты
	if !s.isValidVacationDate(returnDate) {
		msg := tg.NewMessage(update.Message.Chat.ID, "Дата выхода на работу должна быть не раньше завтрашнего дня и не позже чем через 3 недели")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	// Получаем список пользователей для поиска имени
	allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
		return
	}

	// Поиск имени пользователя
	var username string
	for userNameFromEnv, uid := range allUsers {
		if uid == state.SelectedUID {

			username = userNameFromEnv

			break
		}
	}

	if username == "" {
		logger.Instance.Error("Не удалось найти пользователя", "user_id", state.SelectedUID)
		return
	}

	if err := s.startVacation(state.SelectedUID, returnDate); err != nil {
		logger.Instance.Error("Ошибка сохранения статуса отпуска", "error", err)
		return
	}

	message := fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> добавлен в отпуск до %s",
		state.SelectedUID,
		username,
		update.Message.Text)
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
	msg.ReplyToMessageID = update.Message.MessageID

	// Сбрасываем состояние админ-панели
	s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), AdminPanelState{State: AdminStateNone})

	bot.Send(msg)
}

// Обрабатывает команду /admin
func (s *ServiceVacation) handleAdminCommand(update tg.Update, bot *tg.BotAPI) bool {
	// Проверяем, является ли пользователь админом
	if !access.IsAdmin(update.Message.From.ID) {
		msg := tg.NewMessage(update.Message.Chat.ID, "Эта команда доступна только администраторам")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return true
	}

	// Получаем список всех пользователей
	allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
		return true
	}

	keyboard := s.createUsersKeyboard(allUsers)
	msg := tg.NewMessage(update.Message.Chat.ID, "Выберите пользователя:")
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = update.Message.MessageID

	// Сохраняем состояние админ-панели в бд
	state := AdminPanelState{
		State: AdminStateUserList,
	}
	s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), state)

	_, err = bot.Send(msg)
	if err != nil {
		logger.Instance.Error("Ошибка отправки клавиатуры", "error", err)
	}

	return true
}

func (s *ServiceVacation) handleReturnFromVacation(update tg.Update, bot *tg.BotAPI, userNameFromEnv string) bool {
	if !access.IsAdmin(update.Message.From.ID) {
		return false
	}

	// Получаем всех пользователей
	allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
		return true
	}

	// Ищем пользователя по имени среди тех, кто в отпуске
	foundUserId := s.findUserIdByName(allUsers, userNameFromEnv)

	if foundUserId != 0 {
		s.setUserReturnedFromVacation(foundUserId, userNameFromEnv, update, bot)
		s.handleVacationsCommand(update, bot)
		return true
	}

	return false
}

func (s *ServiceVacation) handleChangeVacation(update tg.Update, bot *tg.BotAPI, userName string) bool {
	if !access.IsAdmin(update.Message.From.ID) {
		return false
	}

	allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
		return true
	}

	foundUserId := s.findUserIdByName(allUsers, userName)
	if foundUserId == 0 {
		return false
	}

	// Показываем календарь для выбора новой даты
	dates := s.generateVacationDates()
	keyboard := s.createDateKeyboard(dates)

	msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Выберите новую дату выхода на работу для %s:", userName))
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = update.Message.MessageID

	// Сохраняем состояние и ID пользователя
	state := AdminPanelState{
		State:       AdminStateSetVacation,
		SelectedUID: foundUserId,
	}
	s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), state)

	bot.Send(msg)
	return true
}

// Обрабатывает команду /vacations
func (s *ServiceVacation) handleVacationsCommand(update tg.Update, bot *tg.BotAPI) bool {
	if !access.IsAdmin(update.Message.From.ID) {
		msg := tg.NewMessage(update.Message.Chat.ID, "Эта команда доступна только администраторам")
		bot.Send(msg)
		return true
	}

	message, err := s.GetVacationsList(update)
	if err != nil {
		logger.Instance.Error("Ошибка при получении списка отпусков", "error", err)
		return true
	}

	// Создаем клавиатуру с кнопками для возврата пользователей
	buttons, err := s.createVacationsListKeyboard()
	if err != nil {
		logger.Instance.Error("Ошибка при создании клавиатуры", "error", err)
		return true
	}

	msg := tg.NewMessage(update.Message.Chat.ID, message)

	if len(buttons) > 0 {
		keyboard := tg.NewReplyKeyboard(buttons...)
		keyboard.OneTimeKeyboard = true
		keyboard.Selective = true
		msg.ReplyMarkup = keyboard
		msg.ReplyToMessageID = update.Message.MessageID
	}

	bot.Send(msg)
	return true
}

func (s *ServiceVacation) setUserReturnedFromVacation(userId int64, userName string, update tg.Update, bot *tg.BotAPI) {
	s.endVacation(userId)

	message := fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> возвращен на работу",
		userId,
		userName)
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
}
