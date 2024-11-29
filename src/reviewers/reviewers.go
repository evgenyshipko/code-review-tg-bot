package reviewers

import (
	"code-review-tg-bot/src/logger"
	"code-review-tg-bot/src/storage"
	"encoding/json"
	"errors"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
	"os"
	"slices"
	"strconv"
)

type ReviewerIds map[string]int64

type GetChatMemberType func(config tg.GetChatMemberConfig) (tg.ChatMember, error)

// TODO: тяжелая функция, тоже можно мемоизовать, НО! с инвалидацией по времени, т.к. может изменяться список учстников
func getChatMembers(chatId int64, getChatMember GetChatMemberType) ([]tg.ChatMember, error) {

	reviewersIdsStr := os.Getenv("REVIEW_PARTICIPANTS_IDS")

	var reviewerIdsMap ReviewerIds
	err := json.Unmarshal([]byte(reviewersIdsStr), &reviewerIdsMap)

	if err != nil {
		return []tg.ChatMember{}, fmt.Errorf("ошибка декода переменной REVIEW_PARTICIPANTS_IDS: %s. %s", reviewersIdsStr, err.Error())
	}

	members := []tg.ChatMember{}

	// RND: распараллелить
	for _, reviewerId := range reviewerIdsMap {
		member, err := getChatMember(tg.GetChatMemberConfig{
			ChatConfigWithUser: tg.ChatConfigWithUser{
				ChatID: chatId,
				UserID: reviewerId,
			},
		})
		if err != nil {
			errStr := fmt.Sprintf("%s chatId:%d, userId: %d", err.Error(), chatId, reviewerId)
			logger.Error(errStr)
			continue
		}

		if slices.Contains([]string{"creator", "member", "administrator"}, member.Status) {
			members = append(members, member)
		}
	}

	if len(members) == 0 {
		return []tg.ChatMember{}, fmt.Errorf("Список участников чата пустой")
	}

	return members, nil
}

type usedMembersType map[int64]bool

type GetReviewerFunc func(chatId int64, authorId int64, changedRowsCount int) ([]tg.ChatMember, error)

type ReviewersStorage map[int64]usedMembersType

func GetReviewersCount(changedRows int) (reviewerCount int) {
	reviewerCount = 2
	if changedRows < 20 {
		reviewerCount = 1
	}
	return reviewerCount
}

func setChatUsedReviewersData(chatId int64, usedMembers *usedMembersType) {
	storage.Set("chat"+strconv.FormatInt(chatId, 10), &usedMembers)
}

func getChatUsedReviewersData(chatId int64) *usedMembersType {
	var usedMembers usedMembersType
	exists := storage.Get("chat"+strconv.FormatInt(chatId, 10), &usedMembers)
	if !exists {
		usedMembers = usedMembersType{}
	}
	return &usedMembers
}

func GetReviewers(chatId int64, authorId int64, reviewerCount int, getChatMember GetChatMemberType) ([]tg.ChatMember, error) {

	chatMembers, err := getChatMembers(chatId, getChatMember)
	if err != nil {
		return []tg.ChatMember{}, err
	}

	// usedMembersType - те юзеры, которых не рассматриваем на ревью
	usedMemberIds := *getChatUsedReviewersData(chatId)

	logger.Debug("LENGTH", "len(chatMembers)-1", len(chatMembers)-1, "len(usedMemberIds)", len(usedMemberIds), "reviewerCount", reviewerCount)

	// если видим, что ревьюверов требуется больше, то сразу сбрасываем usedMemberIds
	if len(chatMembers)-1-len(usedMemberIds) < reviewerCount {
		usedMemberIds = usedMembersType{}
	}

	// vacantMembers - те юзеры, которых рассматриваем на ревью
	vacantMembers := make([]tg.ChatMember, 0, len(chatMembers))
	for _, member := range chatMembers {
		if !usedMemberIds[member.User.ID] && member.User.ID != authorId {
			vacantMembers = append(vacantMembers, member)
		}
	}

	logger.Debug("MEMBERS", "usedMemberIds", usedMemberIds, "vacantMembers", vacantMembers)

	if len(vacantMembers) == 0 {
		return []tg.ChatMember{}, errors.New("Кажется, что в данном чате нет пользователей, которые могут осуществлять ревью")
	}

	// выбираем случайных ревьюверов из свободных
	reviewers := make([]tg.ChatMember, 0, len(vacantMembers))

	for len(reviewers) < reviewerCount && (len(reviewers) != len(vacantMembers) || len(reviewers) == 0) {
		randomMember := vacantMembers[rand.Intn(len(vacantMembers))]

		if usedMemberIds[randomMember.User.ID] {
			continue
		}

		usedMemberIds[randomMember.User.ID] = true

		reviewers = append(reviewers, randomMember)
	}

	setChatUsedReviewersData(chatId, &usedMemberIds)

	return reviewers, nil
}
