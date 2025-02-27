package main

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/mergeRequest"
	"code-review-tg-bot/internal/redis"
	"code-review-tg-bot/internal/reviewers"
	"code-review-tg-bot/internal/utils"
	"fmt"
	"os"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/mvdan/xurls"
)

// initialized before main call
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

func main() {
	// Инициализация Redis
	err := redis.Init()
	if err != nil {
		logger.Instance.Errorw("Ошибка инициализации Redis", "error", err)
		panic(err)
	}

	// Получение последнего коммита
	hash, message, err := utils.GetLastCommitInfo()
	if err != nil {
		logger.Instance.Errorw("Ошибка получения информации о последнем коммите", "error", err)
	} else {
		logger.Instance.Infow("Бот стартовал с последним коммитом", "hash", hash, "message", message)
	}

	BotToken := os.Getenv("BOT_TOKEN")

	err = tg.SetLogger(logger.Instance)
	if err != nil {
		panic(err)
	}

	bot, err := tg.NewBotAPI(BotToken)
	if err != nil {
		logger.Instance.Error(err.Error())
		panic(err)
	}

	bot.Debug = true

	// RND: разобраться что это за настройки и на что влияют
	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// RND как работает цикл и причем тут горутины?
	for update := range updates {
		mainLoopFunc(update, bot)
	}

}

//TODO: валидация енвов при запуске
//TODO: избавиться от переменной GITLAB_DOMAIN?
//TODO: доступ только разрешенным разработчикам (и админам т.е завести админов)
//TODO: реализовать команду отпуска
//TODO: кеширование ручек/истории ревью во внешнем источнике (редис)
//TODO: если ссылка на определденный коммит, то делать ревью только этого коммита
//TODO: сделать чтобы бот проставлял ревьюверов в гитлабе

func mainLoopFunc(update tg.Update, bot *tg.BotAPI) {
	defer func() {
		if r := recover(); r != nil {
			logger.Instance.Error("Паника перехвачена", "error", r)

			sendNewMessage(fmt.Sprintf("Что-то пошло не так: %s", r), bot, update)
		}
	}()

	// движемся дальше только если бота тегнули в сообщении
	if update.Message == nil || !strings.Contains(update.Message.Text, bot.Self.UserName) {
		return
	}

	logger.Instance.Infow(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

	// Проверяем доступ пользователя
	if !access.HasAccess(update.Message.From.ID) {
		sendNewMessage(access.GetAccessDeniedMessage(update.Message.From.UserName), bot, update)
		return
	}

	//RND: разобраться - что за параметр -1
	urls := xurls.Strict.FindAllString(update.Message.Text, -1)

	if len(urls) == 0 {
		sendNewMessage("Необходимо добавить ссылку на merge request", bot, update)
		return
	}

	mergeRequestDataStorage := make([]mergeRequest.DataExtended, 0, len(urls))
	totalRowsChanged := 0

	for _, url := range urls {
		mergeRequestData, err := mergeRequest.GetDataByUrl(url)

		if strings.Count(mergeRequestData.Description, "[ ]") > 1 {
			msg := fmt.Sprintf(" <a href=\"%s\">Чеклист</a> из описания МР-а не пройден (пустым может быть только пункт \"Тесты пройдены\", когда тесты отвалились)", url)
			sendNewMessage(msg, bot, update)
			return
		}

		if err != nil {
			logger.Instance.Error(err.Error())
			sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
			return
		}

		if mergeRequestData.HasConflicts {
			msg := fmt.Sprintf("Для начала нужно пофиксить <a href=\"%s\">конфликты</a>", url)
			sendNewMessage(msg, bot, update)
			return
		}

		mergeRequestDataStorage = append(mergeRequestDataStorage, mergeRequestData)

		mergeRequestRowsChanged := mergeRequestData.Deletions + mergeRequestData.Additions

		maxRows, err := strconv.Atoi(os.Getenv("MAXIMUM_ROWS_CHANGED"))

		if err == nil && mergeRequestRowsChanged > maxRows {
			msg := fmt.Sprintf("В <a href=\"%s\">мр-е</a> слишком много строк (>%d). Нужно разбить МР на несколько частей для нормального восприятия ревьюером", url, maxRows)
			sendNewMessage(msg, bot, update)
			return
		}

		totalRowsChanged += mergeRequestRowsChanged
	}

	reviewersCount := reviewers.GetReviewersCount(totalRowsChanged)
	reviewersList, err := reviewers.GetReviewers(update.Message.Chat.ID, update.Message.From.ID, reviewersCount, bot.GetChatMember)
	if err != nil {
		logger.Instance.Error(err.Error())
		sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
		return
	}

	message := generateMessageText(mergeRequestDataStorage, reviewersList)
	sendNewMessage(message, bot, update)
}

func generateMessageText(data []mergeRequest.DataExtended, reviewerList []tg.ChatMember) string {
	msg := "Требуется ревью"

	for _, dataEntity := range data {

		if dataEntity.CommitHash != "" {
			msg += " коммита:\n" + fmt.Sprintf("<a href=\"%s/diffs?commit_id=%s\">%s</a>\nКоммит: %s",
				dataEntity.Url, dataEntity.CommitHash, dataEntity.Title, dataEntity.CommitHash)
		} else {
			msg += ":\n" + fmt.Sprintf("<a href=\"%s\">%s</a>", dataEntity.Url, dataEntity.Title)
		}

		if dataEntity.Extra != nil {
			msg += "\n" + *dataEntity.Extra
		} else if dataEntity.Additions > 0 || dataEntity.Deletions > 0 {
			msg += "\n" + fmt.Sprintf("Размер: +%d -%d", dataEntity.Additions, dataEntity.Deletions)

		}
	}

	msg += "\nРевьюеры: "
	for _, reviewer := range reviewerList {
		msg += "@" + reviewer.User.UserName + " "
	}
	return msg
}

func sendNewMessage(message string, bot *tg.BotAPI, update tg.Update) {
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := bot.Send(msg)
	if err != nil {
		logger.Instance.Error("Ошибка отправки сообщения", "Сообщение не отправлено", err.Error())
	}
}
