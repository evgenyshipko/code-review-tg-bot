package reviewers

import (
	"encoding/json"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand"
	"os"
	"slices"
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

	// TODO: распараллелить
	for _, reviewerId := range reviewerIdsMap {
		member, err := getChatMember(tg.GetChatMemberConfig{
			ChatConfigWithUser: tg.ChatConfigWithUser{
				ChatID: chatId,
				UserID: reviewerId,
			},
		})
		if err != nil {
			errStr := fmt.Sprintf("%s chatId:%d, userId: %d", err.Error(), chatId, reviewerId)
			log.Printf(errStr)
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

type userIdType int64

type usedMembers map[userIdType]bool

type GetReviewerFunc func(chatId int64, changedRowsCount int) ([]tg.ChatMember, error)

func MakeGetReviewerFuncWithMemo(getChatMember GetChatMemberType) GetReviewerFunc {

	memo := make(map[int64]usedMembers)

	return func(chatId int64, changedRowsCount int) ([]tg.ChatMember, error) {

		reviewerCount := 2
		if changedRowsCount < 20 {
			reviewerCount = 1
		}

		chatMembers, err := getChatMembers(chatId, getChatMember)
		if err != nil {
			return []tg.ChatMember{}, err
		}

		usedMemberIds := memo[chatId]

		vacantMembers := make([]tg.ChatMember, 0, len(chatMembers))

		for _, member := range chatMembers {
			if !usedMemberIds[userIdType(member.User.ID)] {
				vacantMembers = append(vacantMembers, member)
			}
		}

		// переобновляем хранилище, если свободных ревьюверов не хватает
		if len(vacantMembers) < reviewerCount {
			memo[chatId] = map[userIdType]bool{}
			vacantMembers = chatMembers
		}

		// выбираем случайных ревьюверов из свободных
		reviewers := make([]tg.ChatMember, 0, len(vacantMembers))
		for len(reviewers) < reviewerCount {
			randomMember := vacantMembers[rand.Intn(len(vacantMembers))]

			if memo[chatId] == nil {
				memo[chatId] = usedMembers{}
			}

			if memo[chatId][userIdType(randomMember.User.ID)] {
				continue
			}

			memo[chatId][userIdType(randomMember.User.ID)] = true
			reviewers = append(reviewers, randomMember)
		}

		fmt.Println("vacantMembers", vacantMembers)
		fmt.Println("reviewers", reviewers)

		return reviewers, nil
	}
}
