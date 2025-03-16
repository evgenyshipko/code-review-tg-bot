package reviewers

import (
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/storage"
	"code-review-tg-bot/internal/vacation"
	"encoding/json"
	"errors"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"sync"
)

type ReviewerIds map[string]int64

type GetChatMemberType func(config tg.GetChatMemberConfig) (tg.ChatMember, error)

type usedMembersType map[int64]bool

type GetReviewerFunc func(chatId int64, authorId int64, changedRowsCount int) ([]tg.ChatMember, error)

type ReviewersStorage map[int64]usedMembersType

type ReviewersService struct {
	storage         storage.Storage
	vacationService *vacation.ServiceVacation
}

func NewReviewersService(storage storage.Storage, vacationService *vacation.ServiceVacation) *ReviewersService {
	return &ReviewersService{
		storage:         storage,
		vacationService: vacationService,
	}
}

type ChatMemberChanData struct {
	err        error
	chatMember tg.ChatMember
}

// TODO: тяжелая функция, тоже можно мемоизовать, НО! с инвалидацией по времени, т.к. может изменяться список учстников
func getChatMembers(chatId int64, getChatMember GetChatMemberType) ([]tg.ChatMember, error) {

	reviewersIdsStr := os.Getenv("REVIEW_PARTICIPANTS_IDS")

	var reviewerIdsMap ReviewerIds
	err := json.Unmarshal([]byte(reviewersIdsStr), &reviewerIdsMap)

	if err != nil {
		return []tg.ChatMember{}, fmt.Errorf("ошибка декода переменной REVIEW_PARTICIPANTS_IDS: %s. %s", reviewersIdsStr, err.Error())
	}

	members := []tg.ChatMember{}
	mu := sync.Mutex{}

	var wg sync.WaitGroup
	chatMemberChan := make(chan ChatMemberChanData, 10)

	for _, reviewerId := range reviewerIdsMap {
		wg.Add(1) // Увеличиваем счетчик горутин

		go func(userId int64) {
			defer wg.Done() // Уменьшаем счетчик при завершении

			member, err := getChatMember(tg.GetChatMemberConfig{
				ChatConfigWithUser: tg.ChatConfigWithUser{
					ChatID: chatId,
					UserID: userId,
				},
			})

			chatMemberChan <- ChatMemberChanData{err: err, chatMember: member}

		}(reviewerId)

	}

	// Закрытие канала после завершения всех горутин
	go func() {
		wg.Wait()             // Ждём завершения всех горутин
		close(chatMemberChan) // Закрываем канал
	}()

	func() {
		for chatMember := range chatMemberChan {
			if chatMember.err != nil {
				// здесь обычно происходит ошибка когда мы пытаемся запросить пользователя, которого в чате нет
				// это нормальная ситуация - пропускаем
				// TODO: если не такая ошибка (например ошибка сети) - надо бы что-то другое предпринять
				logger.Instance.Debugw("getChatMember", "chatId", chatId, "err", err)
				continue
			}

			if slices.Contains([]string{"creator", "member", "administrator"}, chatMember.chatMember.Status) {
				mu.Lock()
				members = append(members, chatMember.chatMember)
				mu.Unlock()
			}
		}
	}()

	if len(members) == 0 {
		return []tg.ChatMember{}, fmt.Errorf("Список участников чата пустой")
	}

	return members, nil
}

func (rs *ReviewersService) GetReviewersCount(changedRows int) (reviewerCount int) {
	reviewerCount = 2
	if changedRows < 20 {
		reviewerCount = 1
	}
	return reviewerCount
}

func (rs *ReviewersService) setChatUsedReviewersData(chatId int64, usedMembers *usedMembersType) {
	rs.storage.Set("chat"+strconv.FormatInt(chatId, 10), &usedMembers)
}

func (rs *ReviewersService) getChatUsedReviewersData(chatId int64) *usedMembersType {
	var usedMembers usedMembersType
	exists := rs.storage.Get("chat"+strconv.FormatInt(chatId, 10), &usedMembers)
	if !exists {
		usedMembers = usedMembersType{}
	}
	return &usedMembers
}

func (s *ReviewersService) GetReviewers(chatId int64, authorId int64, reviewerCount int, getChatMember GetChatMemberType) ([]tg.ChatMember, error) {
	chatMembers, err := getChatMembers(chatId, getChatMember)
	if err != nil {
		return []tg.ChatMember{}, err
	}

	// usedMembersType - те юзеры, которых не рассматриваем на ревью
	usedMemberIds := *s.getChatUsedReviewersData(chatId)

	logger.Instance.Debugw("LENGTH", "len(chatMembers)-1", len(chatMembers)-1, "len(usedMemberIds)", len(usedMemberIds), "reviewerCount", reviewerCount)

	// если видим, что ревьюверов требуется больше, то сразу сбрасываем usedMemberIds
	if len(chatMembers)-1-len(usedMemberIds) < reviewerCount {
		usedMemberIds = usedMembersType{}
	}

	// vacantMembers - те юзеры, которых рассматриваем на ревью
	vacantMembers := make([]tg.ChatMember, 0, len(chatMembers))
	for _, member := range chatMembers {
		if !usedMemberIds[member.User.ID] && member.User.ID != authorId && !s.vacationService.IsUserOnVacation(member.User.ID) {
			vacantMembers = append(vacantMembers, member)
		}
	}

	logger.Instance.Debugw("MEMBERS", "usedMemberIds", usedMemberIds, "vacantMembers", vacantMembers)

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

	s.setChatUsedReviewersData(chatId, &usedMemberIds)

	return reviewers, nil
}
