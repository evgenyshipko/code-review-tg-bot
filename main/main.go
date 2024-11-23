package main

import (
	"code-review-tg-bot/parser"
	"code-review-tg-bot/requests"
	"code-review-tg-bot/reviewers"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/mvdan/xurls"
	"log"
	"os"
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

	getReviewers := reviewers.MakeGetReviewerFuncWithMemo(bot.GetChatMember, 2)

	for update := range updates {
		mainLoopFunc(update, bot, getReviewers)
	}

}

//TODO: настроить единое логгирование

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

	mergeRequestDataStorage := make([]requests.MergeRequestData, 0, len(urls))

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
	}

	reviewersList, err := reviewerFunc(update.Message.Chat.ID)
	if err != nil {
		sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
		return
	}

	message := generateMessageText(mergeRequestDataStorage, reviewersList)
	sendNewMessage(message, bot, update)
}

func generateMessageText(data []requests.MergeRequestData, reviewerList []tg.ChatMember) string {
	msg := "Требуется ревью:"
	for _, dataEntity := range data {
		msg += "\n" + fmt.Sprintf("<a href=\"%s\">%s</a>", dataEntity.Url, dataEntity.Title)
	}

	msg += "\nРевьюеры: "
	for _, reviewer := range reviewerList {
		msg += "@" + reviewer.User.UserName + " "
	}
	return msg
}

func getMergeRequestDataByUrl(url string) (requests.MergeRequestData, error) {
	projectName, mergeRequestId, parseErr := parser.ParseGitlabURL(url)
	if parseErr != nil {
		return requests.MergeRequestData{}, parseErr
	}

	//TODO: projectId - неизменяемая информация, поэтому надо уметь результат этой ручки мемоизировать
	projectId, err := requests.GetProjectId(projectName)
	if err != nil {
		return requests.MergeRequestData{}, err
	}

	mergeRequestData, err := requests.GetMergeRequestData(projectId, mergeRequestId)
	if err != nil {
		return requests.MergeRequestData{}, err
	}

	return mergeRequestData, nil
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
