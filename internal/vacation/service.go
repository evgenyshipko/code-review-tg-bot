package vacation

import (
	"code-review-tg-bot/internal/storage"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ServiceVacation struct {
	storage storage.Storage
	bot     *tg.BotAPI
}

func NewService(storage storage.Storage, bot *tg.BotAPI) *ServiceVacation {
	return &ServiceVacation{
		storage: storage,
		bot:     bot,
	}
}
