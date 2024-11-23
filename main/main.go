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
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	// TODO: разобраться что это за настройки и на что влияют
	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	rxStrict := xurls.Strict

	getReviewers := reviewers.MakeGetReviewerFuncWithMemo(bot.GetChatMember, 2)

baseCycle:
	for update := range updates {

		//defer func() {
		//	if r := recover(); r != nil {
		//
		//	}
		//}()

		if update.Message == nil {
			continue
		}

		if strings.Contains(update.Message.Text, bot.Self.UserName) {
			fmt.Println("[%s] %s", update.Message.From.UserName, update.Message.Text)

			//TODO: что за параметр -1?
			urls := rxStrict.FindAllString(update.Message.Text, -1)

			fmt.Println("urls", urls)

			if len(urls) == 0 {
				sendNewMessage("Необходимо добавить ссылку на merge request", bot, update)
				continue baseCycle
			}

			globalMsg := "Требуется ревью:"

			for _, url := range urls {
				projectName, mergeRequestId, parseErr := parser.ParseGitlabURL(url)
				if parseErr != nil {
					sendNewMessage(parseErr.Error(), bot, update)
					continue baseCycle
				}

				//TODO: projectId - неизменяемая информация, поэтому надо уметь результат этой ручки мемоизировать
				projectId, err := requests.GetProjectId(projectName)
				fmt.Println("projectId", projectId)
				if err != nil {
					sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
					continue baseCycle
				}

				mergeRequestData, err := requests.GetMergeRequestData(projectId, mergeRequestId)
				if err != nil {
					sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
					continue baseCycle
				}

				if mergeRequestData.HasConflicts {
					msg := fmt.Sprintf("Для начала нужно пофиксить <a href=\"%s\">конфликты</a>", url)
					sendNewMessage(msg, bot, update)
					continue baseCycle
				}

				globalMsg += "\n" + fmt.Sprintf("<a href=\"%s\">%s</a>", url, mergeRequestData.Title)
			}

			reviewersList, err := getReviewers(update.Message.Chat.ID)
			if err != nil {
				sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
				continue baseCycle
			}

			//panic("mother fucker")

			fmt.Println("reviewersList[0].User", reviewersList[0].User)

			globalMsg += "\nРевьюеры: "

			for _, reviewer := range reviewersList {
				globalMsg += "@" + reviewer.User.UserName + " "
			}

			sendNewMessage(globalMsg, bot, update)
		}
	}

}

// обрабатывать глобальнуную панику - приложение не должно падать
// настроить единое логгирование

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
