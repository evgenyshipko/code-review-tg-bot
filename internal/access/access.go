package access

import (
	"code-review-tg-bot/internal/logger"
	"encoding/json"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserIds map[int64]string
type Role string

type UserMaps struct {
	ReviewersIdsMap UserIds
	AdminsIdsMap    UserIds
	TestersIdsMap   UserIds
}

type UserRole string

const (
	Tester   UserRole = "Tester"
	Reviewer UserRole = "Reviewer"
	Admin    UserRole = "Admin"
)

type UserData struct {
	Name  string
	Roles []UserRole
}

type Users map[int64]UserData

func MapUserToRoles() (Users, error) {
	var err error

	ReviewersIdsMap, err := ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка ревьюеров", "error", err)
	}

	AdminsIdsMap, err := ParseUserIds("ADMINS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка администраторов", "error", err)
	}

	TestersIdsMap, err := ParseUserIds("TESTERS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка тестировщиков", "error", err)
	}

	users := Users{}
	addUsers(users, ReviewersIdsMap, Reviewer)
	addUsers(users, AdminsIdsMap, Admin)
	addUsers(users, TestersIdsMap, Tester)
	return users, nil
}

func addUsers(users Users, userIds UserIds, role UserRole) {
	for id, name := range userIds {
		_, ok := users[id]
		// If the key exists
		if ok {
			users[id] = UserData{
				Name:  name,
				Roles: append(users[id].Roles, role),
			}
		} else {
			users[id] = UserData{
				Name:  name,
				Roles: []UserRole{role},
			}
		}
	}
}

func InitUserMaps() (*UserMaps, error) {
	var err error

	ReviewersIdsMap, err := ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка ревьюеров", "error", err)
		return nil, err
	}

	AdminsIdsMap, err := ParseUserIds("ADMINS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка администраторов", "error", err)
		return nil, err
	}

	TestersIdsMap, err := ParseUserIds("TESTERS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка тестировщиков", "error", err)
	}

	return &UserMaps{
		ReviewersIdsMap: ReviewersIdsMap,
		AdminsIdsMap:    AdminsIdsMap,
		TestersIdsMap:   TestersIdsMap,
	}, nil
}

func IsUserHasAccess(msg tg.Message, userMaps *UserMaps) bool {
	return isUserInMap(msg.From.ID, userMaps.ReviewersIdsMap) || isUserInMap(msg.From.ID, userMaps.AdminsIdsMap)
}

func ParseUserIds(envName string) (UserIds, error) {
	var tempUserIds map[string]int64
	err := json.Unmarshal([]byte(os.Getenv(envName)), &tempUserIds)
	if err != nil {
		return UserIds{}, err
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

func IsAdmin(userID int64, maps UserMaps) bool {
	return isUserInMap(userID, maps.AdminsIdsMap)
}

func IsReviewer(userId int64, maps UserMaps) bool {
	return isUserInMap(userId, maps.ReviewersIdsMap)
}

func HasVacationAccess(userId int64, maps UserMaps) bool {
	return IsReviewer(userId, maps)
}

func HasAdminAccess(userId int64, maps UserMaps) bool {
	return IsAdmin(userId, maps)
}

func IsTester(userId int64, maps UserMaps) bool {
	return isUserInMap(userId, maps.TestersIdsMap)
}
