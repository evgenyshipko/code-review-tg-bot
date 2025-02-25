package mergeRequest

import (
	"code-review-tg-bot/internal/file"
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

	stats, err := getStats(projectId, diffs, mergeRequestData.SourceBranch)

	if err != nil {
		logger.Instance.Error("ERROR", "STATS CALCULATION", err.Error())
		return DataExtended{mergeRequestData, requests.MergeRequestStats{}}, nil
	}

	return DataExtended{mergeRequestData, stats}, nil
}

func getStats(projectID int, diffs []requests.MergeRequestDiff, gitBranch string) (requests.MergeRequestStats, error) {

	mergeRequestStats := requests.MergeRequestStats{}
MainDiffsLoop:
	for _, diff := range diffs {

		ignorePaths := os.Getenv("IGNORE_PATHS")
		if ignorePaths != "" {
			ignorePathsSlice := strings.Split(ignorePaths, ",")
			for _, ignorePath := range ignorePathsSlice {
				if strings.Contains(diff.NewPath, ignorePath) {
					continue MainDiffsLoop
				}
			}
		}

		if diff.NewFile {
			filePath := diff.NewPath
			if !file.IsCodeFile(filePath) {
				continue
			}

			file1, err := requests.GetRawFile(projectID, filePath, gitBranch)
			if err != nil {
				logger.Instance.Warnw("getStats - статистика не посчиталась", "err", err)
				return requests.MergeRequestStats{}, err
			}

			fileLength := len(strings.Split(file1, "\n"))

			logger.Instance.Debugw("FILE_LEN", fileLength, "filePath", filePath)

			mergeRequestStats.Additions += fileLength
			continue
		}

		for _, line := range strings.Split(diff.Diff, "\n") {
			if strings.HasPrefix(line, "-") {
				mergeRequestStats.Deletions++
			}
			if strings.HasPrefix(line, "+") {
				mergeRequestStats.Additions++
			}
		}
	}

	if mergeRequestStats.Additions == 0 && mergeRequestStats.Deletions == 0 {
		str := "Размер мр-а не посчитан, т.к. он либо не содержит файлов, либо они все проигнорированы"
		mergeRequestStats.Extra = &str
	}

	return mergeRequestStats, nil
}
