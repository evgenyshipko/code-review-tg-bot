package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// Возвращает информацию о последнем коммите
func GetLastCommitInfo() (string, string, error) {
	// Хэш последнего коммита
	hashCmd := exec.Command("git", "rev-parse", "HEAD")

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

	return hash, message, nil
}
