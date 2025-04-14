package main

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/mergeRequest"
	"code-review-tg-bot/internal/reviewers"
	"code-review-tg-bot/internal/storage"
	"code-review-tg-bot/internal/utils"
	"code-review-tg-bot/internal/vacation"
	"fmt"
	"os"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// initialized before main call
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

func main() {
	// Инициализация хранилища
	storageInstance, err := storage.InitStorage()
	if err != nil {
		logger.Instance.Errorw("Ошибка инициализации хранилища", "error", err)
		os.Exit(1)
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

	if err := setUpBotCommands(bot); err != nil {
		logger.Instance.Error("Ошибка настройки команд бота", "error", err)
		os.Exit(1)
	}

	bot.Debug = true

	// RND: разобраться что это за настройки и на что влияют
	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	vacationService := vacation.NewService(storageInstance, bot)
	reviewersService := reviewers.NewReviewersService(storageInstance, vacationService)
	mergeRequestHandler := mergeRequest.NewMergeRequestService(bot, reviewersService)

	// цикл работает пока канал не закрыт
	for update := range updates {
		mainLoopFunc(update, bot, vacationService, mergeRequestHandler)
	}
}

//TODO: валидация енвов при запуске (в том числе бот сейчас не заводится без ADMINS_IDS)
//TODO: избавиться от переменной GITLAB_DOMAIN?
//TODO: кеширование ручек/истории ревью во внешнем источнике (редис)
//TODO: сделать чтобы бот проставлял ревьюверов в гитлабе
//TODO: предусмотреть возможность передачи множества сервисов в mainLoopFunc

func mainLoopFunc(update tg.Update, bot *tg.BotAPI, vs *vacation.ServiceVacation, mr *mergeRequest.MergeRequestService) {
	defer func() {
		if r := recover(); r != nil {
			logger.Instance.Error("Паника перехвачена", "error", r)
			sendNewMessage(fmt.Sprintf("Что-то пошло не так: %s", r), bot, update)
		}
	}()

	// Проверяем доступ пользователя
	if !access.HasAccess(*update.Message, *bot, vacation.ButtonTextConstants.GetHashMap()) {
		sendNewMessage(access.GetAccessDeniedMessage(update.Message.From.UserName), bot, update)

		return
	}

	// Обработка отпусков
	if vs.HandleUpdate(update, bot) {
		return
	}

	// Обработка команд
	if update.Message.IsCommand() {
		handleDefaultCommands(update, bot, vs)
		return
	}

	// движемся дальше только если бота тегнули в сообщении
	if update.Message == nil || !strings.Contains(update.Message.Text, bot.Self.UserName) {
		return
	}

	logger.Instance.Infow(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

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

// Обработка команд
func handleDefaultCommands(update tg.Update, bot *tg.BotAPI, vacationService *vacation.ServiceVacation) {
	if update.Message == nil {
		return
	}

	// Базовые команды
	switch update.Message.Command() {
	case "start":
		msg := tg.NewMessage(update.Message.Chat.ID, "Выберите команду:")
		msg.ReplyMarkup = vacationService.GetDefaultKeyboard(update.Message.From.ID)
		msg.ReplyToMessageID = update.Message.MessageID

		_, err := bot.Send(msg)
		if err != nil {
			logger.Instance.Error("Ошибка отправки клавиатуры", "error", err)
		}
	}
}

func setUpBotCommands(bot *tg.BotAPI) error {
	// Базовые команды
	commands := []tg.BotCommand{
		{
			Command:     "start",
			Description: "Показать клавиатуру с командами",
		},
	}

	// Добавляем команды для работы с отпусками
	vacationService := vacation.NewService(nil, bot) // nil тк нужны только команды
	commands = append(commands, vacationService.GetCommands()...)

	_, err := bot.Request(tg.NewSetMyCommands(commands...))
	if err != nil {
		return fmt.Errorf("ошибка установки команд бота: %w", err)
	}

	return nil
}
