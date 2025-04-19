package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func (s *VacationService) handleButtonPress(update tg.Update, bot *tg.BotAPI) (executed bool) {

	var userState string
	s.storage.Get(fmt.Sprintf("user_state_%d", update.Message.From.ID), &userState)

	if strings.HasPrefix(update.Message.Text, constants.ButtonTextConstants.ReturnFromVacation) {
		userName := strings.TrimPrefix(update.Message.Text, constants.ButtonTextConstants.ReturnFromVacation)
		return s.handleReturnFromVacation(update, bot, userName)
	}

	if strings.HasPrefix(update.Message.Text, constants.ButtonTextConstants.ChangeVacation) {
		userName := strings.TrimPrefix(update.Message.Text, constants.ButtonTextConstants.ChangeVacation)
		return s.handleChangeVacation(update, bot, userName)
	}

	switch update.Message.Text {
	case constants.ButtonTextConstants.VacationsList:
		if !access.IsAdmin(update.Message.From.ID, *s.usersMap) {
			return false
		}
		return s.handleVacationListCommand(update, bot)
	case constants.ButtonTextConstants.SendOnVacation:
		if !access.IsAdmin(update.Message.From.ID, *s.usersMap) {
			return false
		}
		return s.handleSendOnVacation(update, bot)
	case constants.ButtonTextConstants.TakeVacation, constants.ButtonTextConstants.ChangeVacation:
		// Сбрасываем состояние админ-панели
		s.resetAdminState(update.Message.From.ID)
		s.handleChangeVacationStatus(update, bot)
		return true
	case constants.ButtonTextConstants.ReturnToWork, constants.ButtonTextConstants.Cancel:
		// Сбрасываем все состояния
		s.ResetAllStates(update.Message.From.ID)
		s.handleChangeVacationStatus(update, bot)
		return true
	default:
		// Проверяем, является ли сообщение датой после выбора даты в клавиатуре
		if userState == UserStateSelectingDate && s.isDateFormat(update.Message.Text) {
			// Сбрасываем состояние админ-панели
			s.resetAdminState(update.Message.From.ID)
			s.handleChangeVacationStatus(update, bot)
			return true
		}
	}

	return false
}
