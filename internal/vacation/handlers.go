package vacation

import (
	"code-review-tg-bot/internal/constants"
	"fmt"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *VacationService) GetVacationsList() (string, error) {
	message := "Сотрудники в отпуске:\n\n"
	hasVacations := false

	reviewers := s.usersMap.ReviewersIdsMap

	// Проверяем статус отпуска для каждого пользователя
	for userId, userNameFromEnv := range reviewers {
		if s.IsUserOnVacation(userId) {
			returnDate := s.getVacationReturnDate(userId)
			message += fmt.Sprintf("%s - до %s\n",
				userNameFromEnv,
				returnDate.Format(DateFormatLayout))
			hasVacations = true
		}
	}

	if !hasVacations {
		message = "Сейчас никто не в отпуске"
	}

	return message, nil
}

// Возвращает список команд, связанных с отпусками
func (s *VacationService) GetCommands() []tg.BotCommand {
	commands := []tg.BotCommand{
		{
			Command:     constants.TakeVacation,
			Description: "Уйти в отпуск",
		},
		{
			Command:     constants.ReturnToWork,
			Description: "Вернуться к работе",
		},
		{
			Command:     constants.VacationsList,
			Description: "Показать список отпусков (только для админов)",
		},
		{
			Command:     constants.SendOnVacation,
			Description: "Отправить сотрудника в отпуск (только для админов)",
		},
	}

	if len(s.usersMap.TestersIdsMap) > 0 {
		commands = append(commands, tg.BotCommand{
			Command:     constants.TestTakeVacation,
			Description: "Уйти в отпуск на 1 минуту (тестовый режим)",
		})
	}

	return commands
}

// Сохраняет статус отпуска пользователя с TTL
func (s *VacationService) startVacation(userId int64, returnDate time.Time) error {
	now := time.Now()

	// Для обычного отпуска обнуляем время
	if returnDate.Sub(now) > 24*time.Hour {
		returnDate = time.Date(returnDate.Year(), returnDate.Month(), returnDate.Day(), 0, 0, 0, 0, now.Location())
	}

	ttl := returnDate.Sub(now)

	if ttl <= 0 {
		return fmt.Errorf("некорректная дата возврата из отпуска")
	}

	return s.storage.SetWithTTL(fmt.Sprintf("vacation_%d", userId), returnDate, ttl)
}

func (s *VacationService) endVacation(userId int64) {
	s.storage.Delete(fmt.Sprintf("vacation_%d", userId))
}

func (s *VacationService) HandleUpdate(update tg.Update, bot *tg.BotAPI) bool {

	if update.Message.IsCommand() {
		s.HandleCommand(update, bot)
		return true
	}

	if s.isButtonPress(update.Message.Text) {
		s.handleButtonPress(update, bot)
		return true
	}

	return s.handleAdminUpdate(update, bot)
}

func (s *VacationService) GetAdminPanelState(userId int64) AdminPanelState {
	var adminState AdminPanelState
	hasState := s.storage.Get(fmt.Sprintf("admin_state_%d", userId), &adminState)
	if hasState {
		return adminState
	}
	return AdminPanelState{
		State:  AdminStateNone,
		UserId: userId,
	}
}

func (s *VacationService) handleAdminUpdate(update tg.Update, bot *tg.BotAPI) (executed bool) {

	state := s.GetAdminPanelState(update.Message.From.ID)

	s.ResetAllStates(update.Message.From.ID)

	if update.Message.Text == constants.ButtonTextConstants.Cancel {
		// Сбрасываем состояние админ-панели
		s.resetAdminState(update.Message.From.ID)

		msg := tg.NewMessage(update.Message.Chat.ID, MsgKeyboardClosed)
		msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
		return true
	}

	if state.State == AdminStateUserList && strings.HasPrefix(update.Message.Text, "👤 ") {
		s.handleAdminPickUser(update, bot, state)
		return true
	}

	if state.State == AdminStateSetVacation {
		s.handleAdminSetVacation(update, bot, state)
		return true
	}

	if state.State == AdminStateUserActions && update.Message.Text == constants.ButtonTextConstants.ReturnToWork {
		s.handleAdminReturnToWork(update, bot, state)
		return true
	}

	if state.State == AdminStateUserActions && update.Message.Text == constants.ButtonTextConstants.ChangeVacation {
		s.handleAdminChangeVacation(update, bot, state)
		return true
	}

	return false
}
