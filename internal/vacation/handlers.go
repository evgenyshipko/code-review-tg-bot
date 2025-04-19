package vacation

import (
	"code-review-tg-bot/internal/constants"
	"fmt"
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
			Command:     constants.Rest,
			Description: "Уйти в отпуск",
		},
		{
			Command:     constants.Work,
			Description: "Вернуться к работе",
		},
		{
			Command:     constants.Vacations,
			Description: "Показать список отпусков (только для админов)",
		},
		{
			Command:     constants.Vacations_start,
			Description: "Отправить сотрудника в отпуск (только для админов)",
		},
	}

	if len(s.usersMap.TestersIdsMap) > 0 {
		commands = append(commands, tg.BotCommand{
			Command:     constants.Test_vacation,
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

// Все входящие обновления
func (s *VacationService) HandleUpdate(update tg.Update, bot *tg.BotAPI) bool {

	//msg := update.Message.Text
	//
	//
	//if strings.Contains(msg, constants.Start){
	//
	//}

	handlers := []func(tg.Update, *tg.BotAPI) (executed bool){
		s.handleAdminUpdate,
		s.handleButtonPress,
		s.handleTextCommand,
		s.HandleCommand,
	}

	for _, handler := range handlers {
		if handler(update, bot) {
			return true
		}
	}

	return false
}

func (s *VacationService) handleAdminUpdate(update tg.Update, bot *tg.BotAPI) (executed bool) {

	if update.Message != nil {
		var adminState AdminPanelState
		hasState := s.storage.Get(fmt.Sprintf("admin_state_%d", update.Message.From.ID), &adminState)

		if hasState && adminState.State != AdminStateNone {
			s.HandleAdminPanel(update, bot, adminState)
			return true
		}
	}

	return false

}
