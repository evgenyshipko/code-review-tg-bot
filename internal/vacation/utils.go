package vacation

import (
	"strings"
)

func (s *ServiceVacation) findUserIdByName(users map[string]int64, userName string) int64 {
	userName = strings.TrimSpace(userName)

	for userNameFromEnv, userId := range users {
		if s.IsUserOnVacation(userId) {

			fullName := strings.TrimSpace(userNameFromEnv)
			if fullName == userName {
				return userId
			}
		}
	}

	return 0
}
