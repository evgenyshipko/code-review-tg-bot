package mergeRequest

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mvdan/xurls"
)

// Обрабатывает входящее сообщение с запросом на ревью
//   - Извлекает URL'ы из сообщения
//   - Обрабатывает каждый MR
//   - Получает список ревьюеров
//   - Генерирует и отправляет сообщение
func (s *MergeRequestService) Handle(update tg.Update) error {
	urls := s.extractUrls(update.Message.Text)
	if len(urls) == 0 {
		return s.sendMessage("Необходимо добавить ссылку на merge request", update)
	}

	mergeRequests, totalRows, err := s.processMergeRequests(urls)
	if err != nil {
		return err
	}

	reviewers, err := s.getReviewers(update, totalRows)
	if err != nil {
		return fmt.Errorf("ошибка получения ревьюеров: %w", err)
	}

	message := s.generateMessage(mergeRequests, reviewers)
	return s.sendMessage(message, update)
}

// Извлекает все URLы из текста сообщения
func (s *MergeRequestService) extractUrls(text string) []string {
	return xurls.Strict.FindAllString(text, -1)
}

// Обрабатывает список URLов мерж реквестов
func (s *MergeRequestService) processMergeRequests(urls []string) ([]DataExtended, int, error) {
	var mergeRequests []DataExtended
	totalRows := 0

	for _, url := range urls {
		mr, rows, err := s.processSingleMergeRequest(url)
		if err != nil {
			return nil, 0, err
		}
		mergeRequests = append(mergeRequests, mr)
		totalRows += rows
	}

	return mergeRequests, totalRows, nil
}

// Обрабатывает один мерж реквест
func (s *MergeRequestService) processSingleMergeRequest(url string) (DataExtended, int, error) {
	mr, err := GetDataByUrl(url)
	if err != nil {
		return DataExtended{}, 0, fmt.Errorf("ошибка получения данных MR: %w", err)
	}

	if err := s.validateMergeRequest(mr, url); err != nil {
		return DataExtended{}, 0, err
	}

	rowsChanged := mr.Deletions + mr.Additions
	if err := s.validateRowsCount(rowsChanged, url); err != nil {
		return DataExtended{}, 0, err
	}

	return mr, rowsChanged, nil
}

func (s *MergeRequestService) validateMergeRequest(mr DataExtended, url string) error {
	if strings.Count(mr.Description, "[ ]") > 1 {
		return fmt.Errorf("<a href=\"%s\">Чеклист</a> из описания МР-а не пройден (пустым может быть только пункт \"Тесты пройдены\", когда тесты отвалились)", url)
	}

	if mr.HasConflicts {
		return fmt.Errorf("Для начала нужно пофиксить <a href=\"%s\">конфликты</a>", url)
	}

	return nil
}

func (s *MergeRequestService) validateRowsCount(rowsChanged int, url string) error {
	maxRows, err := strconv.Atoi(os.Getenv("MAXIMUM_ROWS_CHANGED"))
	if err != nil {
		return nil // Если не удалось получить максимальное количество строк, пропускаем валидацию
	}

	if rowsChanged > maxRows {
		return fmt.Errorf("В <a href=\"%s\">мр-е</a> слишком много строк (>%d). Нужно разбить МР на несколько частей для нормального восприятия ревьюером", url, maxRows)
	}

	return nil
}

func (s *MergeRequestService) getReviewers(update tg.Update, totalRows int) ([]tg.ChatMember, error) {
	reviewersCount := s.reviewersService.GetReviewersCount(totalRows)
	return s.reviewersService.GetReviewers(
		update.Message.Chat.ID,
		update.Message.From.ID,
		reviewersCount,
		s.bot.GetChatMember,
	)
}

func (s *MergeRequestService) generateMessage(mrs []DataExtended, reviewers []tg.ChatMember) string {
	msg := "Требуется ревью"

	for _, mr := range mrs {
		if mr.CommitHash != "" {
			msg += " коммита:\n" + fmt.Sprintf("<a href=\"%s/diffs?commit_id=%s\">%s</a>\nКоммит: %s",
				mr.Url, mr.CommitHash, mr.Title, mr.CommitHash)
		} else {
			msg += ":\n" + fmt.Sprintf("<a href=\"%s\">%s</a>", mr.Url, mr.Title)
		}

		if mr.Extra != nil {
			msg += "\n" + *mr.Extra
		} else if mr.Additions > 0 || mr.Deletions > 0 {
			msg += "\n" + fmt.Sprintf("Размер: +%d -%d", mr.Additions, mr.Deletions)
		}
	}

	msg += "\nРевьюеры: "
	for _, reviewer := range reviewers {
		msg += "@" + reviewer.User.UserName + " "
	}
	return msg
}

func (s *MergeRequestService) sendMessage(message string, update tg.Update) error {
	msg := tg.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = tg.ModeHTML
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := s.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	return nil
}
