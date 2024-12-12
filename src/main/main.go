package main

import (
	"code-review-tg-bot/src/logger"
	"code-review-tg-bot/src/mergeRequest"
	"code-review-tg-bot/src/reviewers"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/mvdan/xurls"
	"os"
	"strconv"
	"strings"
)

// initialized before main call
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

func main() {

	BotToken := os.Getenv("BOT_TOKEN")

	err := tg.SetLogger(logger.Logger)
	if err != nil {
		panic(err)
	}

	bot, err := tg.NewBotAPI(BotToken)
	if err != nil {
		logger.Error(err.Error())
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

func mainLoopFunc(update tg.Update, bot *tg.BotAPI) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Паника перехвачена", "error", r, "stack", logger.GetStackTraceAsSlice())

			sendNewMessage(fmt.Sprintf("Что-то пошло не так: %s", r), bot, update)
		}
	}()

	// движемся дальше только если бота тегнули в сообщении
	if update.Message == nil || !strings.Contains(update.Message.Text, bot.Self.UserName) {
		return
	}

	logger.Info(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

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
		if err != nil {
			logger.Error(err.Error())
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
		logger.Error(err.Error())
		sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
		return
	}

	message := generateMessageText(mergeRequestDataStorage, reviewersList)
	sendNewMessage(message, bot, update)
}

func generateMessageText(data []mergeRequest.DataExtended, reviewerList []tg.ChatMember) string {
	msg := "Требуется ревью:"
	for _, dataEntity := range data {
		msg += "\n" + fmt.Sprintf("<a href=\"%s\">%s</a>", dataEntity.Url, dataEntity.Title)
		if dataEntity.Additions > 0 && dataEntity.Deletions > 0 {
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
		logger.Error("Ошибка отправки сообщения", "Сообщение не отправлено", err.Error())
	}
}
