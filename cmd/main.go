package main

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/constants"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/mergeRequest"
	"code-review-tg-bot/internal/reviewers"
	"code-review-tg-bot/internal/storage"
	"code-review-tg-bot/internal/utils"
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

	userMaps, err := access.InitUserMaps()
	if err != nil {
		logger.Instance.Errorw("Ошибка инициализации списков пользователей", "error", err)
		os.Exit(1)
	}

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

	vacationService := vacation.NewVacationService(storageInstance, bot, userMaps)
	reviewersService := reviewers.NewReviewersService(storageInstance, vacationService)
	mergeRequestHandler := mergeRequest.NewMergeRequestService(bot, reviewersService)

	if err := setUpBotCommands(bot, vacationService.GetCommands()); err != nil {
		logger.Instance.Error("Ошибка настройки команд бота", "error", err)
		os.Exit(1)
	}

	bot.Debug = true

	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// ЗАПОМНИТЬ: цикл работает пока канал не закрыт
	for update := range updates {
		mainLoopFunc(update, bot, vacationService, mergeRequestHandler, &users)
	}
}

//TODO: валидация енвов при запуске (в том числе бот сейчас не заводится без ADMINS_IDS)
//TODO: избавиться от переменной GITLAB_DOMAIN?
//TODO: кеширование ручек/истории ревью во внешнем источнике (редис)
//TODO: сделать чтобы бот проставлял ревьюверов в гитлабе
//TODO: предусмотреть возможность передачи множества сервисов в mainLoopFunc

func mainLoopFunc(update tg.Update, bot *tg.BotAPI, vs *vacation.VacationService, mr *mergeRequest.MergeRequestService, users *access.Users) {
	defer func() {
		if r := recover(); r != nil {
			logger.Instance.Error("Паника перехвачена", "error", r)
			sendNewMessage(fmt.Sprintf("Что-то пошло не так: %s", r), bot, update)
		}
	}()

	if !access.IsUserHasAccess(update.Message.From.ID, *users) {
		return
	}

	logger.Instance.Infow(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

	if vs.HandleUpdate(update, bot) {
		return
	}

	if !strings.Contains(update.Message.Text, "@"+bot.Self.UserName) {
		return
	}

	if err := mr.Handle(update); err != nil {
		logger.Instance.Error(err.Error())
		sendNewMessage("Что-то пошло не так: "+err.Error(), bot, update)
	}
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

func setUpBotCommands(bot *tg.BotAPI, vacationCommands []tg.BotCommand) error {
	// Базовые команды
	commands := []tg.BotCommand{
		{
			Command:     constants.Start,
			Description: "Показать клавиатуру с командами",
		},
	}

	commands = append(commands, vacationCommands...)

	_, err := bot.Request(tg.NewSetMyCommands(commands...))
	if err != nil {
		return fmt.Errorf("ошибка установки команд бота: %w", err)
	}

	return nil
}
