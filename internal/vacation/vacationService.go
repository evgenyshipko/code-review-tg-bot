package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/storage"
	"fmt"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type VacationService struct {
	storage     storage.Storage
	bot         *tg.BotAPI
	users       *access.Users
	ReviewerIds *access.UserIds
}

func NewVacationService(storage storage.Storage, bot *tg.BotAPI, reviewerIds *access.UserIds, users *access.Users) *VacationService {
	return &VacationService{
		storage:     storage,
		bot:         bot,
		ReviewerIds: reviewerIds,
		users:       users,
	}
}

func (s *VacationService) IsUserOnVacation(userId int64) bool {
	exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &time.Time{})
	return exists
}

func (s *VacationService) GetVacationReturnDate(userId int64) time.Time {
	var returnDate time.Time
	s.storage.Get(fmt.Sprintf("vacation_%d", userId), &returnDate)
	return returnDate
}

func (s *VacationService) SetUserReturnedFromVacation(userId int64, userName string, update tg.Update, bot *tg.BotAPI) {
	s.EndVacation(userId)

	message := fmt.Sprintf("Пользователь <a href=\"tg://user?id=%d\">%s</a> возвращен на работу",
		userId,
		userName)
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyMarkup = tg.NewRemoveKeyboard(true)
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
}

func (s *VacationService) FindVacationUserIdByName(users access.UserIds, userName string) int64 {
	userName = strings.TrimSpace(userName)

	for userId, userNameFromEnv := range users {
		if s.IsUserOnVacation(userId) {

			fullName := strings.TrimSpace(userNameFromEnv)
			if fullName == userName {
				return userId
			}
		}
	}

	return 0
}

type UserData struct {
	Name       string
	FinishDate time.Time
}

func (s *VacationService) GetUsersInVacation() []UserData {
	var users []UserData

	for userId, userNameFromEnv := range *s.ReviewerIds {
		if s.IsUserOnVacation(userId) {
			returnDate := s.GetVacationReturnDate(userId)
			users = append(users, UserData{
				Name:       userNameFromEnv,
				FinishDate: returnDate,
			})
		}
	}
	return users
}

func (s *VacationService) GetUsersInVacationMessage(users []UserData) string {
	message := "Сотрудники в отпуске:\n\n"
	for _, user := range users {
		message += fmt.Sprintf("%s - до %s\n",
			user.Name,
			user.FinishDate.Format(constants.DateFormatLayout))
	}
	return message
}

func (s *VacationService) StartVacation(userId int64, returnDate time.Time) error {
	now := time.Now()

	returnDate = time.Date(returnDate.Year(), returnDate.Month(), returnDate.Day(), 0, 0, 0, 0, now.Location())

	ttl := returnDate.Sub(now)

	if ttl <= 0 {
		return fmt.Errorf("некорректная дата возврата из отпуска")
	}

	return s.storage.SetWithTTL(fmt.Sprintf("vacation_%d", userId), returnDate, ttl)
}

func (s *VacationService) EndVacation(userId int64) {
	s.storage.Delete(fmt.Sprintf("vacation_%d", userId))
}
