package utils

import (
	"code-review-tg-bot/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Возвращает информацию о последнем коммите
func GetLastCommitInfo() (string, string, error) {
	// Хэш последнего коммита
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	lastCommitHashFromEnv := os.Getenv("LAST_COMMIT_HASH")

	hashOutput, err := hashCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting commit hash: %v", err)
	}
	hash := strings.TrimSpace(string(hashOutput))

	// Сообщение последнего коммита
	msgCmd := exec.Command("git", "log", "-1", "--pretty=%B")
	msgOutput, err := msgCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting commit message: %v", err)
	}
	message := strings.TrimSpace(string(msgOutput))

	logger.Instance.Infow("Получение последнего коммита из переменных окружения", "lastCommitHashFromEnv", lastCommitHashFromEnv)
	if hashCmd.String() == lastCommitHashFromEnv {
		logger.Instance.Infow("Хэши коммитов идентичны", "hash:", hash, "lastCommitHashFromEnv: ", lastCommitHashFromEnv)
	}

	return hash, message, nil
}
