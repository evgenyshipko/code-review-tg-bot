package vacation

import (
	"fmt"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *ServiceVacation) findUserIdByName(users map[string]int64, chatID int64, userName string) int64 {
	userName = strings.TrimSpace(userName)

	for _, userId := range users {
		var status StatusVacationUser
		exists := s.storage.Get(fmt.Sprintf("vacation_%d", userId), &status)

		if exists && status.IsOnVacation {
			member, err := s.bot.GetChatMember(tg.GetChatMemberConfig{
				ChatConfigWithUser: tg.ChatConfigWithUser{
					ChatID: chatID,
					UserID: userId,
				},
			})
			if err != nil {
				continue
			}

			fullName := strings.TrimSpace(member.User.FirstName + " " + member.User.LastName)
			if fullName == userName {
				return userId
			}
		}
	}

	return 0
}
