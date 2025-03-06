package vacation

import (
	"code-review-tg-bot/internal/storage"
	"fmt"

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
	var status StatusVacationUser
	exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &status)
	return exists && status.IsOnVacation
}

// Cбрасывает состояние пользователя
func (s *ServiceVacation) ResetUserState(userId int64) {
	s.storage.Set(fmt.Sprintf("user_state_%d", userId), UserStateNone)
}

// Cбрасывает состояние админ-панели
func (s *ServiceVacation) ResetAdminState(userId int64) {
	s.storage.Set(fmt.Sprintf("admin_state_%d", userId), AdminPanelState{State: AdminStateNone})
}

// Cбрасывает все состояния пользователя
func (s *ServiceVacation) ResetAllStates(userId int64) {
	s.ResetUserState(userId)
	s.ResetAdminState(userId)
}
