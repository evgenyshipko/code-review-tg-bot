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
	if update.Message.Text == ButtonCancel {
		// Сбрасываем состояние админ-панели
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), AdminPanelState{State: AdminStateNone})

		msg := tg.NewMessage(update.Message.Chat.ID, "Админ-панель закрыта")
		msg.ReplyMarkup = s.GetDefaultKeyboard()
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

// Возвращает список пользователей в отпуске, доступно только админам
func (s *ServiceVacation) GetVacationsList(update tg.Update) (string, error) {
	message := "Сотрудники в отпуске:\n\n"
	hasVacations := false

	// Получаем всех пользователей из env
	allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		return "", fmt.Errorf("ошибка при парсинге списка пользователей: %w", err)
	}

	// Проверяем статус отпуска для каждого пользователя
	for _, userId := range allUsers {
		var status StatusVacationUser

		member, err := s.bot.GetChatMember(tg.GetChatMemberConfig{
			ChatConfigWithUser: tg.ChatConfigWithUser{
				ChatID: update.Message.Chat.ID,
				UserID: userId,
			},
		})

		if err != nil {
			continue
		}

		exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &status)

		if exists && status.IsOnVacation {
			message += fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a> - до %s\n",
				userId,
				member.User.FirstName+" "+member.User.LastName,
				status.ReturnDate.Format(DateFormatLayout))
			hasVacations = true
		}
	}

	if !hasVacations {
		message = "Сейчас никто не в отпуске"
	}

	return message, nil
}

// Возвращает клавиатуру по умолчанию
func (s *ServiceVacation) GetDefaultKeyboard() tg.ReplyKeyboardMarkup {
	var defaultButtons = []tg.KeyboardButton{
		tg.NewKeyboardButton(ButtonTakeVacation),
		tg.NewKeyboardButton(ButtonReturnToWork),
	}

	return tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(defaultButtons...),
	)
}

// Обрабатывает текстовые команды, связанные с отпуском
func (s *ServiceVacation) HandleTextCommand(update tg.Update, bot *tg.BotAPI) bool {
	var commandMap = map[string]string{
		"отпуск": ButtonTakeVacation,
		"работа": ButtonReturnToWork,
	}
	text := strings.ToLower(update.Message.Text)

	for keyword, action := range commandMap {
		if strings.Contains(text, keyword) {
			s.handleStatus(tg.Update{
				Message: &tg.Message{
					Text: action,
					From: update.Message.From,
					Chat: update.Message.Chat,
				},
			}, bot)
			return true
		}

	}

	return false
}

// Возвращает список команд, связанных с отпусками
func (s *ServiceVacation) GetCommands() []tg.BotCommand {
	return []tg.BotCommand{
		{
			Command:     "rest",
			Description: "Уйти в отпуск",
		},
		{
			Command:     "work",
			Description: "Вернуться к работе",
		},
		{
			Command:     "vacations",
			Description: "Показать список отпусков (только для админов)",
		},
		{
			Command:     "admin",
			Description: "Открыть админ-панель (только для админов)",
		},
	}
}

// Обрабатывает все команды, связанные с отпусками и админ-панелью
func (s *ServiceVacation) HandleCommand(update tg.Update, bot *tg.BotAPI) bool {
	if update.Message == nil {
		return false
	}

	switch update.Message.Command() {
	case "admin":
		return s.handleAdminCommand(update, bot)
	case "rest":
		s.handleStatus(tg.Update{
			Message: &tg.Message{
				Text: ButtonTakeVacation,
				From: update.Message.From,
				Chat: update.Message.Chat,
			},
		}, bot)
		return true
	case "work":
		s.handleStatus(tg.Update{
			Message: &tg.Message{
				Text: ButtonReturnToWork,
				From: update.Message.From,
				Chat: update.Message.Chat,
			},
		}, bot)
		return true
	case "vacations":
		return s.handleVacationsCommand(update, bot)
	}

	return false
}

// Обрабатывает нажатия на кнопки, возвращает true, если обработка прошла успешно
func (s *ServiceVacation) HandleButtonPress(update tg.Update, bot *tg.BotAPI) bool {
	if update.Message == nil {
		return false
	}

	switch update.Message.Text {
	case ButtonTakeVacation, ButtonReturnToWork, ButtonCancel:
		s.handleStatus(update, bot)
		return true
	default:
		// Проверяем, является ли сообщение датой после выбора даты в клавиатуре
		if s.isDateFormat(update.Message.Text) {
			s.handleStatus(update, bot)
			return true
		}
	}

	return false
}

// Внутренние методы

// Обрабатывает изменение статуса отпуска
func (s *ServiceVacation) handleStatus(update tg.Update, bot *tg.BotAPI) {
	switch update.Message.Text {
	case ButtonTakeVacation:
		// Показываем календарь на клавиатуре
		dates := s.generateVacationDates()
		keyboard := s.createDateKeyboard(dates)

		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите дату выхода на работу:")
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки календаря", "error", err)
		}

	case ButtonReturnToWork:
		status := StatusVacationUser{
			IsOnVacation: false,
		}
		s.storage.Set(fmt.Sprintf("vacation_%d", update.Message.From.ID), status)

		message := fmt.Sprintf("@%s вернулся к работе", update.Message.From.UserName)
		msg := tg.NewMessage(update.Message.Chat.ID, message)
		msg.ReplyMarkup = s.GetDefaultKeyboard()
		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки статуса", "error", err)
		}

	case ButtonCancel:
		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите статус:")
		msg.ReplyMarkup = s.GetDefaultKeyboard()
		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки клавиатуры", "error", err)
		}

	default:
		// Проверяем, является ли сообщение датой
		returnDate, err := time.Parse(DateFormatLayout, update.Message.Text)
		if err == nil {
			status := StatusVacationUser{
				IsOnVacation: true,
				ReturnDate:   returnDate,
			}

			// Обновляем статус в бд
			s.storage.Set(fmt.Sprintf("vacation_%d", update.Message.From.ID), status)

			message := fmt.Sprintf("@%s ушёл в отпуск до %s", update.Message.From.UserName, update.Message.Text)
			msg := tg.NewMessage(update.Message.Chat.ID, message)
			msg.ReplyMarkup = s.GetDefaultKeyboard()
			_, err := bot.Send(msg)

			if err != nil {
				logger.Instance.Error("Ошибка отправки статуса", "error", err)
			}
		}
	}
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

	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	bot.Send(msg)

	return true
}

// Обрабатывает команду /admin
func (s *ServiceVacation) handleAdminCommand(update tg.Update, bot *tg.BotAPI) bool {
	// Проверяем, является ли пользователь админом
	if !access.IsAdmin(update.Message.From.ID) {
		msg := tg.NewMessage(update.Message.Chat.ID, "Эта команда доступна только администраторам")
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

// Обрабатывает выбор пользователя
func (s *ServiceVacation) handleAdminPickUser(update tg.Update, bot *tg.BotAPI, state AdminPanelState) {
	if strings.HasPrefix(update.Message.Text, "👤 ") {
		username := strings.TrimPrefix(update.Message.Text, "👤 ")

		// Получаем ID пользователя по имени
		allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
		if err != nil {
			logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
			return
		}

		// Получаем ID выбранного пользователя
		var selectedUserId int64
		for uname, uid := range allUsers {
			if uname == username {
				selectedUserId = uid
				break
			}
		}

		if selectedUserId == 0 {
			msg := tg.NewMessage(update.Message.Chat.ID, "Пользователь не найден")
			bot.Send(msg)
			return
		}

		// Создаем клавиатуру с действиями
		keyboard := tg.NewReplyKeyboard(
			tg.NewKeyboardButtonRow(
				tg.NewKeyboardButton("📅 Добавить в отпуск"),
				tg.NewKeyboardButton("💼 Вернуть на работу"),
			),
			tg.NewKeyboardButtonRow(
				tg.NewKeyboardButton(ButtonCancel),
			),
		)
		keyboard.OneTimeKeyboard = true

		msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Выберите действие для пользователя %s:", username))
		msg.ReplyMarkup = keyboard

		// Обновляем состояние
		state.State = AdminStateUserActions
		state.SelectedUID = selectedUserId
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), state)

		bot.Send(msg)
	}
}

// Обрабатывает действия в админ-панели
func (s *ServiceVacation) handleAdminActionsForUser(update tg.Update, bot *tg.BotAPI, state AdminPanelState) {
	switch update.Message.Text {
	case "📅 Добавить в отпуск":
		// Показываем календарь
		dates := s.generateVacationDates()
		keyboard := s.createDateKeyboard(dates)

		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите дату выхода на работу:")
		msg.ReplyMarkup = keyboard

		// Обновляем состояние
		state.State = AdminStateSetVacation
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), state)

		bot.Send(msg)

	case "💼 Вернуть на работу":
		// Получаем список пользователей для поиска имени
		allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
		if err != nil {
			logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
			return
		}

		// Ищем имя пользователя по ID
		var username string
		for _, uid := range allUsers {
			if uid == state.SelectedUID {
				member, err := s.bot.GetChatMember(tg.GetChatMemberConfig{
					ChatConfigWithUser: tg.ChatConfigWithUser{
						ChatID: update.Message.Chat.ID,
						UserID: state.SelectedUID,
					},
				})
				if err != nil {
					logger.Instance.Error("Ошибка при получении информации о пользователе", "error", err)
					return
				}

				username = member.User.FirstName + " " + member.User.LastName
				break
			}
		}

		if username == "" {
			logger.Instance.Error("Не удалось найти имя пользователя", "user_id", state.SelectedUID)
			return
		}

		// Устанавливаем статус "не в отпуске"
		status := StatusVacationUser{
			IsOnVacation: false,
		}
		s.storage.Set(fmt.Sprintf("vacation_%d", state.SelectedUID), status)

		message := fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> возвращен на работу",
			state.SelectedUID,
			username)
		msg := tg.NewMessage(update.Message.Chat.ID, message)
		msg.ParseMode = tg.ModeHTML
		msg.ReplyMarkup = s.GetDefaultKeyboard()

		// Сбрасываем состояние админ-панели
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), AdminPanelState{State: AdminStateNone})

		bot.Send(msg)
	}
}

// Обрабатывает добавление пользователя в отпуск
func (s *ServiceVacation) handleAdminSetVacation(update tg.Update, bot *tg.BotAPI, state AdminPanelState) {
	// Проверка, является ли сообщение датой
	returnDate, err := time.Parse(DateFormatLayout, update.Message.Text)

	if err == nil {
		// Получаем список пользователей для поиска имени
		allUsers, err := access.ParseUserIds("REVIEW_PARTICIPANTS_IDS")
		if err != nil {
			logger.Instance.Error("Ошибка при парсинге списка пользователей", "error", err)
			return
		}

		// Поиск имени пользователя
		var username string
		for _, uid := range allUsers {
			if uid == state.SelectedUID {
				member, err := s.bot.GetChatMember(tg.GetChatMemberConfig{
					ChatConfigWithUser: tg.ChatConfigWithUser{
						ChatID: update.Message.Chat.ID,
						UserID: state.SelectedUID,
					},
				})

				if err != nil {
					logger.Instance.Error("Ошибка при получении информации о пользователе", "error", err)
					return
				}

				username = member.User.FirstName + " " + member.User.LastName

				break
			}
		}

		if username == "" {
			logger.Instance.Error("Не удалось найти пользователя", "user_id", state.SelectedUID)
			return
		}

		// Сохраняем отпуск пользователя в бд
		status := StatusVacationUser{
			IsOnVacation: true,
			ReturnDate:   returnDate,
		}
		s.storage.Set(fmt.Sprintf("vacation_%d", state.SelectedUID), status)

		message := fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> добавлен в отпуск до %s",
			state.SelectedUID,
			username,
			update.Message.Text)
		msg := tg.NewMessage(update.Message.Chat.ID, message)
		msg.ParseMode = tg.ModeHTML
		msg.ReplyMarkup = s.GetDefaultKeyboard()

		// Сбрасываем состояние админ-панели
		s.storage.Set(fmt.Sprintf("admin_state_%d", update.Message.From.ID), AdminPanelState{State: AdminStateNone})

		bot.Send(msg)
	}
}

// createUsersKeyboard создает клавиатуру со списком пользователей
func (s *ServiceVacation) createUsersKeyboard(users map[string]int64) tg.ReplyKeyboardMarkup {
	var rows [][]tg.KeyboardButton
	for username := range users {
		rows = append(rows, []tg.KeyboardButton{
			tg.NewKeyboardButton(fmt.Sprintf("👤 %s", username)),
		})
	}

	// Добавляем кнопку отмены
	rows = append(rows, []tg.KeyboardButton{
		tg.NewKeyboardButton(ButtonCancel),
	})

	keyboard := tg.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true

	return keyboard
}

// Генерирует даты выхода на работу
func (s *ServiceVacation) generateVacationDates() []string {
	var dates []string
	now := time.Now()
	vacationLimitDays := 21

	for i := 1; i <= vacationLimitDays; i++ {
		date := now.AddDate(0, 0, i)
		dates = append(dates, date.Format(DateFormatLayout))
	}

	return dates
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

	return tg.NewReplyKeyboard(rows...)
}

// Проверяет, является ли текст датой в валидном формате
func (s *ServiceVacation) isDateFormat(text string) bool {
	_, err := time.Parse(DateFormatLayout, text)
	return err == nil
}
