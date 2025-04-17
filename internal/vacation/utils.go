package vacation

import (
	"code-review-tg-bot/internal/access"
	"strings"
)

func (s *VacationService) findUserIdByName(users access.UserIds, userName string) int64 {
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
