package vacation

import (
	"code-review-tg-bot/internal/storage"
	"fmt"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ServiceVacation struct {
	storage storage.Storage
	bot     *tg.BotAPI
}

func NewService(storage storage.Storage, bot *tg.BotAPI) *ServiceVacation {
	return &ServiceVacation{
		storage: storage,
		bot:     bot,
	}
}

// Gроверяет, находится ли пользователь в отпуске
func (s *ServiceVacation) IsUserOnVacation(userId int64) bool {
	exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &time.Time{})
	return exists
}

// Cбрасывает состояние пользователя
func (s *ServiceVacation) resetUserState(userId int64) {
	s.storage.Set(fmt.Sprintf("user_state_%d", userId), UserStateNone)
}

// Cбрасывает состояние админ-панели
func (s *ServiceVacation) resetAdminState(userId int64) {
	s.storage.Set(fmt.Sprintf("admin_state_%d", userId), AdminPanelState{State: AdminStateNone})
}

// Cбрасывает все состояния пользователя
func (s *ServiceVacation) ResetAllStates(userId int64) {
	s.resetUserState(userId)
	s.resetAdminState(userId)
}

// Получает дату возвращения из отпуска
func (s *ServiceVacation) getVacationReturnDate(userId int64) time.Time {
	var returnDate time.Time
	s.storage.Get(fmt.Sprintf("vacation_%d", userId), &returnDate)
	return returnDate
}
