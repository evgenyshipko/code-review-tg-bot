package mergeRequest

import (
	"code-review-tg-bot/src/file"
	"code-review-tg-bot/src/logger"
	"code-review-tg-bot/src/parser"
	"code-review-tg-bot/src/requests"
	"os"
	"strconv"
	"strings"
)

type DataExtended struct {
	requests.MergeRequestData
	requests.MergeRequestStats
}

func GetDataByUrl(url string) (DataExtended, error) {
	projectName, mergeRequestId, parseErr := parser.ParseGitlabURL(url)
	if parseErr != nil {
		return DataExtended{}, parseErr
	}

	//TODO: projectId - неизменяемая информация, поэтому надо уметь результат этой ручки мемоизировать
	projectId, err := requests.GetProjectId(projectName)
	if err != nil {
		return DataExtended{}, err
	}

	mergeRequestData, err := requests.GetMergeRequestData(projectId, mergeRequestId)
	if err != nil {
		return DataExtended{}, err
	}

	filesCount, err := strconv.Atoi(mergeRequestData.ChangesCount)
	if err != nil {
		filesCount = 20
	}

	stats, err := getStats(projectId, mergeRequestId, filesCount, mergeRequestData.SourceBranch)
	if err != nil {
		return DataExtended{mergeRequestData, requests.MergeRequestStats{}}, nil
	}

	return DataExtended{mergeRequestData, stats}, nil
}

func getStats(projectID int, mergeRequestId int, filesCount int, gitBranch string) (requests.MergeRequestStats, error) {
	diffs, err := requests.GetMergeRequestDiffs(projectID, mergeRequestId, filesCount)
	if err != nil {
		return requests.MergeRequestStats{}, err
	}

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
				logger.Error("ERROR WHEN GET RAW FILE", err.Error())
				return requests.MergeRequestStats{}, err
			}

			fileLength := len(strings.Split(file1, "\n"))

			logger.Debug("FILE_LEN", fileLength, "filePath", filePath)

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

	return mergeRequestStats, nil
}
