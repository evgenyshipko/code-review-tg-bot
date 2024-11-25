package main

import (
	"code-review-tg-bot/src/parser"
	"code-review-tg-bot/src/requests"
	"code-review-tg-bot/src/reviewers"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/mvdan/xurls"
	"log"
	"os"
	"strconv"
	"strings"
)

// initialized before main call
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {

	BotToken := os.Getenv("BOT_TOKEN")

	bot, err := tg.NewBotAPI(BotToken)
	// TODO: законсолить нормально ошибку
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	// TODO: разобраться что это за настройки и на что влияют
	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	getReviewers := reviewers.MakeGetReviewerFuncWithMemo(bot.GetChatMember)

	for update := range updates {
		mainLoopFunc(update, bot, getReviewers)
	}

}

//TODO: настроить единое логгирование
//TODO: валидация енвов при запуске
//TODO: избавиться от переменной GITLAB_DOMAIN?

func mainLoopFunc(update tg.Update, bot *tg.BotAPI, reviewerFunc reviewers.GetReviewerFunc) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in mainLoopFunc", r)
			sendNewMessage(fmt.Sprintf("Что-то пошло не так: %s", r), bot, update)
		}
	}()

	// движемся дальше только если бота тегнули в сообщении
	if update.Message == nil || !strings.Contains(update.Message.Text, bot.Self.UserName) {
		return
	}

	fmt.Println("[%s] %s", update.Message.From.UserName, update.Message.Text)

	//TODO: что за параметр -1?
	urls := xurls.Strict.FindAllString(update.Message.Text, -1)

	if len(urls) == 0 {
		sendNewMessage("Необходимо добавить ссылку на merge request", bot, update)
		return
	}

	mergeRequestDataStorage := make([]MergeRequestDataExtended, 0, len(urls))
	totalRowsChanged := 0

	for _, url := range urls {
		mergeRequestData, err := getMergeRequestDataByUrl(url)
		if err != nil {
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

	reviewersList, err := reviewerFunc(update.Message.Chat.ID, totalRowsChanged)
	if err != nil {
		sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
		return
	}

	message := generateMessageText(mergeRequestDataStorage, reviewersList)
	sendNewMessage(message, bot, update)
}

func generateMessageText(data []MergeRequestDataExtended, reviewerList []tg.ChatMember) string {
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

type MergeRequestDataExtended struct {
	requests.MergeRequestData
	requests.MergeRequestStats
}

func getMergeRequestDataByUrl(url string) (MergeRequestDataExtended, error) {
	projectName, mergeRequestId, parseErr := parser.ParseGitlabURL(url)
	if parseErr != nil {
		return MergeRequestDataExtended{}, parseErr
	}

	//TODO: projectId - неизменяемая информация, поэтому надо уметь результат этой ручки мемоизировать
	projectId, err := requests.GetProjectId(projectName)
	if err != nil {
		return MergeRequestDataExtended{}, err
	}

	mergeRequestData, err := requests.GetMergeRequestData(projectId, mergeRequestId)
	if err != nil {
		return MergeRequestDataExtended{}, err
	}

	stats, err := GetMergeRequestStats(projectId, mergeRequestId)
	if err != nil {
		return MergeRequestDataExtended{mergeRequestData, requests.MergeRequestStats{}}, err
	}

	return MergeRequestDataExtended{mergeRequestData, stats}, nil
}

func GetMergeRequestStats(projectID int, mergeRequestId int) (requests.MergeRequestStats, error) {
	commits, err := requests.GetMergeRequestCommits(projectID, mergeRequestId)
	if err != nil {
		return requests.MergeRequestStats{}, err
	}
	mergeRequestStats := requests.MergeRequestStats{}
	for _, commit := range commits {
		commitStats, err := requests.GetCommitData(projectID, commit.ID)
		if err != nil {
			return requests.MergeRequestStats{}, err
		}
		mergeRequestStats.Additions += commitStats.Stats.Additions
		mergeRequestStats.Deletions += commitStats.Stats.Deletions
	}
	return mergeRequestStats, nil
}

func sendNewMessage(message string, bot *tg.BotAPI, update tg.Update) {
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := bot.Send(msg)
	if err != nil {
		fmt.Println("Сообщение не отправлено", err)
		sendNewMessage("Не смог отправить сообщение", bot, update)
	}
}
