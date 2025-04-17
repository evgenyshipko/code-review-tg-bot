package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/storage"
	"fmt"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type VacationService struct {
	storage  storage.Storage
	bot      *tg.BotAPI
	usersMap *access.UserMaps
}

func NewVacationService(storage storage.Storage, bot *tg.BotAPI, usersMap *access.UserMaps) *VacationService {
	return &VacationService{
		storage:  storage,
		bot:      bot,
		usersMap: usersMap,
	}
}

// Gроверяет, находится ли пользователь в отпуске
func (s *VacationService) IsUserOnVacation(userId int64) bool {
	exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &time.Time{})
	return exists
}

// Cбрасывает состояние пользователя
func (s *VacationService) resetUserState(userId int64) {
	s.storage.Set(fmt.Sprintf("user_state_%d", userId), UserStateNone)
}

// Cбрасывает состояние админ-панели
func (s *VacationService) resetAdminState(userId int64) {
	s.storage.Set(fmt.Sprintf("admin_state_%d", userId), AdminPanelState{State: AdminStateNone})
}

// Cбрасывает все состояния пользователя
func (s *VacationService) ResetAllStates(userId int64) {
	s.resetUserState(userId)
	s.resetAdminState(userId)
}

// Получает дату возвращения из отпуска
func (s *VacationService) getVacationReturnDate(userId int64) time.Time {
	var returnDate time.Time
	s.storage.Get(fmt.Sprintf("vacation_%d", userId), &returnDate)
	return returnDate
}
