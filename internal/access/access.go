package access

import (
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/logger"
	"encoding/json"
	"os"
	"slices"
)

type UserIds map[int64]string
type Role string

type UserMaps struct {
	ReviewersIdsMap UserIds
	AdminsIdsMap    UserIds
	TestersIdsMap   UserIds
}

type UserData struct {
	Name  string
	Roles []constants.UserRole
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
	addUsers(users, ReviewersIdsMap, constants.Reviewer)
	addUsers(users, AdminsIdsMap, constants.Admin)
	addUsers(users, TestersIdsMap, constants.Tester)
	return users, nil
}

func addUsers(users Users, userIds UserIds, role constants.UserRole) {
	for id, name := range userIds {
		_, ok := users[id]
		if ok {
			users[id] = UserData{
				Name:  name,
				Roles: append(users[id].Roles, role),
			}
		} else {
			users[id] = UserData{
				Name:  name,
				Roles: []constants.UserRole{role},
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

func IsUserHasAccess(userId int64, users Users) bool {
	if val, ok := users[userId]; ok {
		for _, role := range val.Roles {
			if role == constants.Reviewer || role == constants.Admin {
				return true
			}
		}
	}
	return false
}

func IsUserHasAccessToCommand(userId int64, command string, users Users) bool {
	availableRoles := CommandToRoleMapping[command]
	usersRoles := users[userId].Roles
	for _, role := range usersRoles {
		if slices.Contains(availableRoles, role) {
			return true
		}
	}
	return false
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

var CommandToRoleMapping = map[string][]constants.UserRole{
	constants.Start:           []constants.UserRole{constants.Admin, constants.Tester, constants.Reviewer},
	constants.Rest:            []constants.UserRole{constants.Reviewer},
	constants.Work:            []constants.UserRole{constants.Reviewer},
	constants.Vacations:       []constants.UserRole{constants.Admin},
	constants.Vacations_start: []constants.UserRole{constants.Admin},
	constants.Test_vacation:   []constants.UserRole{constants.Tester, constants.Admin},
}
