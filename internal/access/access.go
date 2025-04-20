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

	users := Users{}
	addUsers(users, ReviewersIdsMap, constants.Reviewer)
	addUsers(users, AdminsIdsMap, constants.Admin)
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

func GetReviewersMap() (*UserIds, error) {
	var err error

	ReviewersIdsMap, err := ParseUserIds("REVIEW_PARTICIPANTS_IDS")
	if err != nil {
		logger.Instance.Error("Ошибка при парсинге списка ревьюеров", "error", err)
		return nil, err
	}

	return &ReviewersIdsMap, nil
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

func IsUserHasAccessToCommand(userId int64, command constants.BotCommand, users Users) bool {
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

var CommandToRoleMapping = map[constants.BotCommand][]constants.UserRole{
	constants.Start:            []constants.UserRole{constants.Admin, constants.Reviewer},
	constants.TakeVacation:     []constants.UserRole{constants.Reviewer},
	constants.ReturnToWork:     []constants.UserRole{constants.Reviewer},
	constants.VacationsList:    []constants.UserRole{constants.Admin},
	constants.SendToVacation:   []constants.UserRole{constants.Admin},
	constants.TestTakeVacation: []constants.UserRole{constants.Admin},
}
