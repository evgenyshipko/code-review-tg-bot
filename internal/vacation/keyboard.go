package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackActions string

const (
	PickUser               CallbackActions = "pick_user"
	CancelUser             CallbackActions = "cancel_user"
	PickDate               CallbackActions = "pick_date"
	CancelDate             CallbackActions = "cancel_date"
	UserReturnFromVacation CallbackActions = "user_return_from_vacation"
	UserChangeVacation     CallbackActions = "user_change_vacation"
	CancelVacationList     CallbackActions = "cancel_vacation_list"
	ReturnToWork           CallbackActions = "return_to_work"
	TakeVacation           CallbackActions = "take_vacation"
	VacationList           CallbackActions = "employees_list"
	SendToVacation         CallbackActions = "send_to_vacation"
	Cancel                 CallbackActions = "cancel"
)

func (s *VacationService) createUsersKeyboard(users access.UserIds) tg.InlineKeyboardMarkup {
	var rows [][]tg.InlineKeyboardButton

	for userId, userNameFromEnv := range users {
		// Пропускаем пользователей, которые уже в отпуске
		if s.IsUserOnVacation(userId) {
			continue
		}

		rows = append(rows, []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData(fmt.Sprintf("👤 %s", userNameFromEnv), string(PickUser)),
		})
	}

	rows = append(rows, []tg.InlineKeyboardButton{
		tg.NewInlineKeyboardButtonData(constants.ButtonTextConstants.Cancel, string(CancelUser)),
	})

	keyboard := tg.NewInlineKeyboardMarkup(rows...)

	return keyboard
}

func (s *VacationService) createDateKeyboard(dates []string) tg.InlineKeyboardMarkup {
	var rows [][]tg.InlineKeyboardButton
	var elemCountInRow int = 3

	// Создаем ряды по 3 кнопки в каждом
	for i := 0; i < len(dates); i += elemCountInRow {
		var row []tg.InlineKeyboardButton
		for j := 0; j < elemCountInRow && i+j < len(dates); j++ {
			row = append(row, tg.NewInlineKeyboardButtonData(dates[i+j], string(PickDate)))
		}
		rows = append(rows, row)
	}

	rows = append(rows, []tg.InlineKeyboardButton{
		tg.NewInlineKeyboardButtonData(constants.ButtonTextConstants.Cancel, string(CancelDate)),
	})

	keyboard := tg.NewInlineKeyboardMarkup(rows...)
	return keyboard
}

func (s *VacationService) createVacationsListKeyboard() ([][]tg.InlineKeyboardButton, error) {
	var buttons [][]tg.InlineKeyboardButton

	allUsers := s.usersMap.ReviewersIdsMap

	for userId, userName := range allUsers {
		if s.IsUserOnVacation(userId) {
			buttons = append(buttons,
				[]tg.InlineKeyboardButton{
					tg.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", constants.ButtonTextConstants.ReturnFromVacation, userName), string(UserReturnFromVacation)),
					tg.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", constants.ButtonTextConstants.ChangeVacation, userName), string(UserChangeVacation)),
				},
			)
		}
	}

	if len(buttons) > 0 {
		buttons = append(buttons, []tg.InlineKeyboardButton{tg.NewInlineKeyboardButtonData(constants.ButtonTextConstants.Cancel, string(CancelVacationList))})
	}

	return buttons, nil
}

func (s *VacationService) getDefaultKeyboard(userId int64) tg.InlineKeyboardMarkup {
	var defaultButtons []tg.InlineKeyboardButton

	// Показываем кнопки отпуска только ревьюерам
	if access.HasVacationAccess(userId, *s.usersMap) {

		if s.IsUserOnVacation(userId) {
			defaultButtons = append(defaultButtons, tg.NewInlineKeyboardButtonData(constants.ButtonTextConstants.ReturnToWork, string(ReturnToWork)))
		} else {
			defaultButtons = append(defaultButtons, tg.NewInlineKeyboardButtonData(constants.ButtonTextConstants.TakeVacation, string(TakeVacation)))
		}
	}

	// Показываем админ-кнопки только админам
	if access.HasAdminAccess(userId, *s.usersMap) {
		defaultButtons = append(defaultButtons,
			tg.NewInlineKeyboardButtonData("📋 Список отпусков", string(VacationList)),
			tg.NewInlineKeyboardButtonData("➕ Отправить в отпуск", string(SendToVacation)))
	}

	keyboard := tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(defaultButtons...),
		tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonData(constants.ButtonTextConstants.Cancel, string(CancelDate))),
	)
	return keyboard
}
