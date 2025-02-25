package mergeRequest

import (
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/parser"
	"code-review-tg-bot/internal/requests"
	"os"
	"strings"
)

type DataExtended struct {
	requests.MergeRequestData
	requests.MergeRequestStats
}

func GetDataByUrl(url string) (DataExtended, error) {
	encodedQueryPathProject, mergeRequestId, commitHash, parseErr := parser.ParseGitlabURL(url)
	if parseErr != nil {
		return DataExtended{}, parseErr
	}

	//TODO: projectId - неизменяемая информация, поэтому надо уметь результат этой ручки мемоизировать
	projectId, err := requests.GetProjectId(encodedQueryPathProject)
	if err != nil {
		return DataExtended{}, err
	}

	mergeRequestData, err := requests.GetMergeRequestData(projectId, mergeRequestId, commitHash)
	if err != nil {
		return DataExtended{}, err
	}

	var diffs []requests.MergeRequestDiff
	if commitHash != "" {
		diffs, err = requests.GetCommitDiffs(projectId, commitHash)
	} else {
		diffs, err = requests.GetMergeRequestDiffs(projectId, mergeRequestId)

	}

	if err != nil {
		return DataExtended{}, err
	}

	stats, err := getStats(diffs)

	if err != nil {
		logger.Instance.Error("ERROR", "STATS CALCULATION", err.Error())
		return DataExtended{mergeRequestData, requests.MergeRequestStats{}}, nil
	}

	return DataExtended{mergeRequestData, stats}, nil
}

// Вычисляет статистику мр-а
func getStats(diffs []requests.MergeRequestDiff) (requests.MergeRequestStats, error) {
	mergeRequestStats := requests.MergeRequestStats{}
	ignoredPaths := getIgnoredPaths()

	for _, diff := range diffs {
		if shouldIgnorePath(diff.NewPath, ignoredPaths) {
			continue
		}

		additions, deletions := countAdditionsAndDeletionsRows(diff.Diff)
		mergeRequestStats.Additions += additions
		mergeRequestStats.Deletions += deletions
	}

	if mergeRequestStats.Additions == 0 && mergeRequestStats.Deletions == 0 {
		message := "Размер мр-а не посчитан, т.к. он либо не содержит файлов, либо они все проигнорированы"
		mergeRequestStats.Extra = &message
	}

	return mergeRequestStats, nil
}

// Загружает список путей из IGNORE_PATH для игнорирования
func getIgnoredPaths() []string {
	ignorePaths := os.Getenv("IGNORE_PATHS")
	if ignorePaths == "" {
		return nil
	}

	return strings.Split(ignorePaths, ",")
}

// Проверяет, нужно ли игнорировать путь
func shouldIgnorePath(path string, ignoredPaths []string) bool {
	for _, ignorePath := range ignoredPaths {
		if strings.Contains(path, ignorePath) {
			return true
		}
	}

	return false
}

// Подсчитывает добавленные/удаленные строки
func countAdditionsAndDeletionsRows(diff string) (int, int) {
	var additions, deletions int

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") {
			additions++
		}
		if strings.HasPrefix(line, "-") {
			deletions++
		}
	}

	return additions, deletions
}
