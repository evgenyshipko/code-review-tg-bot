package vacation

import (
	"code-review-tg-bot/internal/access"
	"fmt"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
			message += fmt.Sprintf("%s - до %s\n",
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
			Command:     "vacations_leave",
			Description: "Отправить сотрудника в отпуск (только для админов)",
		},
	}
}

// Сохраняет статус отпуска пользователя с TTL
func (s *ServiceVacation) saveVacationStatus(userId int64, status StatusVacationUser) error {
	if !status.IsOnVacation {
		s.storage.Set(fmt.Sprintf("vacation_%d", userId), status)
		return nil
	}

	now := time.Now()
	returnDate := time.Date(status.ReturnDate.Year(), status.ReturnDate.Month(), status.ReturnDate.Day(), 0, 0, 0, 0, now.Location())
	ttl := returnDate.Sub(now)

	if ttl <= 0 {
		return fmt.Errorf("некорректная дата возврата из отпуска")
	}

	return s.storage.SetWithTTL(fmt.Sprintf("vacation_%d", userId), status, ttl)
}
