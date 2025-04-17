package permissions

import (
	"code-review-tg-bot/internal/vacation"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func IsAllowedMessage(msg tgbotapi.Message, botName string, reviewersNames []string) bool {

	isReviewerName := false
	for _, reviewerName := range reviewersNames {
		if strings.Contains(msg.Text, reviewerName) {
			isReviewerName = true
		}
	}

	return msg.IsCommand() ||
		strings.Contains(msg.Text, "@"+botName) ||
		vacation.ButtonTextConstants.GetHashMap()[msg.Text] || isReviewerName
}
