package mergeRequest

import (
	"code-review-tg-bot/internal/reviewers"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MergeRequestService struct {
	bot              *tg.BotAPI
	reviewersService *reviewers.ReviewersService
}

func NewMergeRequestService(bot *tg.BotAPI, rs *reviewers.ReviewersService) *MergeRequestService {
	return &MergeRequestService{
		bot:              bot,
		reviewersService: rs,
	}
}
