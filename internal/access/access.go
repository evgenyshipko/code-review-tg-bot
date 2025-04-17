package access

import (
	"code-review-tg-bot/internal/logger"

	"encoding/json"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserIds map[int64]string
type Role string

const (
	AccessRole    Role = "accessRole"
	NotAccessRole Role = "notAccessRole"
)

var (
	ReviewersIdsMap UserIds
	AdminsIdsMap    UserIds
	TestersIdsMap   UserIds
)

func InitUserMaps() error {
	var err error

	ReviewersIdsMap, err = ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка ревьюеров", "error", err)
		return err
	}

	AdminsIdsMap, err = ParseUserIds("ADMINS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка администраторов", "error", err)
		return err
	}

	TestersIdsMap, err = ParseUserIds("TESTERS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка тестировщиков", "error", err)
		return err
	}

	return nil
}

func HasAccessByRole(msg tg.Message) (Role, error) {
	if isUserInMap(msg.From.ID, ReviewersIdsMap) || isUserInMap(msg.From.ID, AdminsIdsMap) {
		return AccessRole, nil
	}
	return NotAccessRole, nil
}

func ParseUserIds(envName string) (UserIds, error) {
	var tempUserIds map[string]int64
	err := json.Unmarshal([]byte(os.Getenv(envName)), &tempUserIds)
	if err != nil {
		return nil, err
	}

	userIds := make(UserIds)
	for name, id := range tempUserIds {
		userIds[id] = name
	}

	return userIds, nil
}

func isUserInMap(userId int64, userMap UserIds) bool {
	_, exists := userMap[userId]
	return exists
}

func IsAdmin(userID int64) bool {
	return isUserInMap(userID, AdminsIdsMap)
}

func IsReviewer(userId int64) bool {
	return isUserInMap(userId, ReviewersIdsMap)
}

func HasVacationAccess(userId int64) bool {
	return IsReviewer(userId)
}

func HasAdminAccess(userId int64) bool {
	return IsAdmin(userId)
}

func IsTester(userId int64) bool {
	return isUserInMap(userId, TestersIdsMap)
}
