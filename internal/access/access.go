package access

import (
	"code-review-tg-bot/internal/logger"
	"encoding/json"
	"fmt"
	"os"
)

type UserIds map[string]int64

// Проверяет, имеет ли пользователь с указанным ID доступ к боту
func HasAccess(userId int64) bool {
	reviewersIdsMap, err := parseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка ревьюеров", "error", err)
		return false
	}

	adminsIdsMap, err := parseUserIds("ADMINS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка администраторов", "error", err)
		return false
	}

	return isUserInMap(userId, reviewersIdsMap) || isUserInMap(userId, adminsIdsMap)
}

// Парсит список пользователей из .env
func parseUserIds(envName string) (UserIds, error) {
	var userIds UserIds
	err := json.Unmarshal([]byte(os.Getenv(envName)), &userIds)

	return userIds, err
}

// Проверяет, есть ли пользователь в мапе
func isUserInMap(userId int64, userMap UserIds) bool {
	for _, id := range userMap {
		if id == userId {
			return true
		}
	}
	return false
}

// Формирует сообщение об отказе в доступе
func GetAccessDeniedMessage(username string) string {
	return fmt.Sprintf("@%s у вас нет доступа к этому боту", username)
}

// ParseUserIds парсит список пользователей из .env
func ParseUserIds(envName string) (UserIds, error) {
	var userIds UserIds
	err := json.Unmarshal([]byte(os.Getenv(envName)), &userIds)

	return userIds, err
}

// IsUserInMap проверяет, есть ли пользователь в мапе
func IsUserInMap(userId int64, userMap UserIds) bool {
	for _, id := range userMap {
		if id == userId {
			return true
		}
	}
	return false
}

// Проверяет, является ли пользователь администратором
func IsAdmin(userID int64) bool {
	adminsIdsMap, err := ParseUserIds("ADMINS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка администраторов", "error", err)
		return false
	}

	return IsUserInMap(userID, adminsIdsMap)
}

// Проверяет, является ли пользователь ревьюером
func IsReviewer(userId int64) bool {
	reviewers, err := ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		return false
	}

	for _, id := range reviewers {
		if id == userId {
			return true
		}
	}
	return false
}

// Проверяет, имеет ли пользователь доступ к функциям отпуска
func HasVacationAccess(userId int64) bool {
	return IsReviewer(userId)
}

// Проверяет, имеет ли пользователь доступ к админ-функциям
func HasAdminAccess(userId int64) bool {
	return IsAdmin(userId)
}

// Проверяет, является ли пользователь тестировщиком
func IsTester(userId int64) bool {
	testers, err := ParseUserIds("TESTERS_IDS")
	if err != nil {
		return false
	}

	return IsUserInMap(userId, testers)
}
