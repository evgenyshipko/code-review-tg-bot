package main

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/mergeRequest"
	"code-review-tg-bot/internal/reviewers"
	"code-review-tg-bot/internal/storage"
	"code-review-tg-bot/internal/stories"
	"code-review-tg-bot/internal/vacation"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

func main() {
	BotToken := os.Getenv("BOT_TOKEN")

	err := tg.SetLogger(logger.Instance)
	if err != nil {
		panic(err)
	}

	bot, err := tg.NewBotAPI(BotToken)
	if err != nil {
		logger.Instance.Error(err.Error())
		panic(err)
	}

	mergeRequestService, storyService := InitServices(bot)
	
	if err := setUpBotCommands(bot); err != nil {
		logger.Instance.Error("Ошибка настройки команд бота", "error", err)
		os.Exit(1)
	}

	bot.Debug = true

	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// ЗАПОМНИТЬ: цикл работает пока канал не закрыт
	for update := range updates {
		userInputHandler(update, bot, mergeRequestService, storyService)
	}
}

func InitServices(bot *tg.BotAPI) (*mergeRequest.MergeRequestService, *stories.StoryService) {
	storageInstance, err := storage.InitStorage()
	if err != nil {
		logger.Instance.Errorw("Ошибка инициализации хранилища", "error", err)
		os.Exit(1)
	}

	users, err := access.MapUserToRoles()
	if err != nil {
		logger.Instance.Warnw("access.MapUserToRoles", "error", err)
	}
	logger.Instance.Info(users)

	reviewersMap, err := access.GetReviewersMap()
	if err != nil {
		logger.Instance.Errorw("Ошибка инициализации списков пользователей", "error", err)
		os.Exit(1)
	}

	err = tg.SetLogger(logger.Instance)
	if err != nil {
		panic(err)
	}

	vacationService := vacation.NewVacationService(storageInstance, bot, reviewersMap, &users)
	reviewersService := reviewers.NewReviewersService(storageInstance, vacationService)
	mergeRequestService := mergeRequest.NewMergeRequestService(bot, reviewersService)

	storiesArr := []stories.Story{*stories.TakeVacationStory, *stories.ReturnToWorkStory, *stories.ShowVacationListStory, *stories.SendToVacationStory}
	storyService := stories.NewStoryService(storageInstance, storiesArr, vacationService, bot, &users, reviewersMap)
	return mergeRequestService, storyService
}

//TODO: валидация енвов при запуске (в том числе бот сейчас не заводится без ADMINS_IDS)
//TODO: избавиться от переменной GITLAB_DOMAIN?
//TODO: кеширование ручек/истории ревью во внешнем источнике (редис)
//TODO: сделать чтобы бот проставлял ревьюверов в гитлабе

// TODO: вынести из main
func userInputHandler(update tg.Update, bot *tg.BotAPI, mr *mergeRequest.MergeRequestService, storyService *stories.StoryService) {
	defer func() {
		if r := recover(); r != nil {
			logger.Instance.Error("Паника перехвачена", "error", r)
			sendNewMessage(fmt.Sprintf("Что-то пошло не так: %s", r), bot, update)
		}
	}()

	if !access.IsUserHasAccess(update.Message.From.ID, *storyService.Users) {
		return
	}

	userId := update.Message.From.ID
	message := update.Message

	if message != nil {
		logger.Instance.Infow(fmt.Sprintf("[%s] %s", message.From.UserName, message.Text))
	}

	// Если у пользователя есть активная история, то работаем в ее рамках
	executed := storyService.HandleCurrentStories(update, userId)
	if executed {
		return
	}

	if message.IsCommand() {
		command := constants.BotCommand(message.Command())
		if !access.IsUserHasAccessToCommand(update.Message.From.ID, command, *storyService.Users) {
			msg := tg.NewMessage(update.Message.Chat.ID, "У Вас нет доступа к данной команде")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return
		}

		switch command {
		case constants.TakeVacation:
			storyService.ExecuteStory(stories.TakeVacationStory, nil, update, userId)
		case constants.ReturnToWork:
			storyService.ExecuteStory(stories.ReturnToWorkStory, nil, update, userId)
		case constants.VacationsList:
			storyService.ExecuteStory(stories.ShowVacationListStory, nil, update, userId)
		case constants.SendToVacation:
			storyService.ExecuteStory(stories.SendToVacationStory, nil, update, userId)
		}
		return
	}

	if !strings.Contains(message.Text, "@"+bot.Self.UserName) {
		return
	}

	if err := mr.Handle(update); err != nil {
		logger.Instance.Error(err.Error())
		sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
	}
}

// TODO: вынести из main и прокидывать методом в сервисы
func sendNewMessage(message string, bot *tg.BotAPI, update tg.Update) {
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := bot.Send(msg)
	if err != nil {
		logger.Instance.Error("Ошибка отправки сообщения", "Сообщение не отправлено", err.Error())
	}
}

// TODO: вынести из main
func setUpBotCommands(bot *tg.BotAPI) error {
	_, err := bot.Request(tg.NewSetMyCommands(constants.BotCommands...))
	if err != nil {
		return fmt.Errorf("ошибка установки команд бота: %w", err)
	}

	return nil
}
