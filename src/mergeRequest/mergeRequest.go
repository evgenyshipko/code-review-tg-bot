package mergeRequest

import (
	"code-review-tg-bot/src/parser"
	"code-review-tg-bot/src/requests"
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

	stats, err := getStats(projectId, mergeRequestId)
	if err != nil {
		return DataExtended{mergeRequestData, requests.MergeRequestStats{}}, err
	}

	return DataExtended{mergeRequestData, stats}, nil
}

func getStats(projectID int, mergeRequestId int) (requests.MergeRequestStats, error) {
	diffs, err := requests.GetMergeRequestDiffs(projectID, mergeRequestId)
	if err != nil {
		return requests.MergeRequestStats{}, err
	}
	mergeRequestStats := requests.MergeRequestStats{}
	for _, diff := range diffs {
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
