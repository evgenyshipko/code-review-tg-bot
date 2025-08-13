package vacation

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *VacationService) CreateUsersKeyboard(users access.UserIds) tg.ReplyKeyboardMarkup {
	var rows [][]tg.KeyboardButton

	for userId, userNameFromEnv := range users {
		if s.IsUserOnVacation(userId) {
			continue
		}

		rows = append(rows, []tg.KeyboardButton{
			//TODO: избавиться от логики на префиксах
			tg.NewKeyboardButton(fmt.Sprintf("👤 %s", userNameFromEnv)),
		})
	}

	rows = append(rows, []tg.KeyboardButton{
		tg.NewKeyboardButton(constants.ButtonTextConstants.Cancel),
	})

	keyboard := tg.NewReplyKeyboard(rows...)

	return keyboard
}

func CreateDateKeyboard(dates []string) tg.ReplyKeyboardMarkup {
	var rows [][]tg.KeyboardButton
	var elemCountInRow int = 3

	for i := 0; i < len(dates); i += elemCountInRow {
		var row []tg.KeyboardButton
		for j := 0; j < elemCountInRow && i+j < len(dates); j++ {
			row = append(row, tg.NewKeyboardButton(dates[i+j]))
		}
		rows = append(rows, row)
	}

	rows = append(rows, []tg.KeyboardButton{
		tg.NewKeyboardButton(constants.ButtonTextConstants.Cancel),
	})

	keyboard := tg.NewReplyKeyboard(rows...)
	return keyboard
}

func (s *VacationService) CreateVacationsListKeyboard() ([][]tg.KeyboardButton, error) {
	var buttons [][]tg.KeyboardButton

	for userId, userName := range *s.ReviewerIds {
		if s.IsUserOnVacation(userId) {
			buttons = append(buttons,
				[]tg.KeyboardButton{
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", constants.ButtonTextConstants.ReturnFromVacation, userName)),
					tg.NewKeyboardButton(fmt.Sprintf("%s %s", constants.ButtonTextConstants.ChangeVacation, userName)),
				},
			)
		}
	}

	if len(buttons) > 0 {
		buttons = append(buttons, []tg.KeyboardButton{tg.NewKeyboardButton(constants.ButtonTextConstants.Cancel)})
	}

	return buttons, nil
}
