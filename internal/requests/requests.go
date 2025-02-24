package requests

import (
	"code-review-tg-bot/internal/logger"
	"fmt"
	"net/url"
	"os"

	"github.com/go-resty/resty/v2"
)

type ProjectData struct {
	ID int `json:"id"`
}

type MergeRequestData struct {
	Title        string `json:"title"`
	HasConflicts bool   `json:"has_conflicts"`
	Description  string `json:"description"`
	Url          string `json:"web_url"`
	ChangesCount string `json:"changes_count"`
	SourceBranch string `json:"source_branch"`
	CommitHash   string `json:"-"`
}

type HTTPClient struct {
	*resty.Client
}

func NewHTTPClient() *resty.Client {
	return resty.New()
}

/*
GET запрос к апи гитлаба
  - path: путь к эндпоинту
  - result: указатель для сохранения респонса
  - isRawResponse: флаг для приведение респонса к строке
*/
func doGitlabGet(path string, result interface{}, args ...bool) error {

	logResponse := true
	if len(args) > 0 {
		logResponse = args[0]
	}

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")

	fullURL := fmt.Sprintf("https://%s/api/v4/%s", gitlabDomain, path)

	resp, err :=
		NewHTTPClient().
			R().
			SetHeader("PRIVATE-TOKEN", gitlabToken).
			SetResult(&result).
			Get(fullURL)

	strRes, ok := result.(*string)
	if ok {
		*strRes = resp.String()
	}

	if err != nil {
		return fmt.Errorf("не удалось выполнить GET-запрос: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("выполнен запрос с ошибкой, код: %d, тело ответа: %s",
			resp.StatusCode(), resp.String())
	}

	responseReasult := result
	if !logResponse {
		responseReasult = "NO LOGS"
	}

	logger.Instance.Infow(
		"doGitlabGet",
		"pathRequest: ", path,
		"responseResult: ", responseReasult)

	return nil
}

func GetProjectId(projectPathName string) (int, error) {
	var data ProjectData
	path := fmt.Sprintf("projects/%s", projectPathName)

	if err := doGitlabGet(path, &data); err != nil {
		return 0, err
	}

	return data.ID, nil
}

func GetMergeRequestData(projectID int, mergeRequestId int, commitHash string) (MergeRequestData, error) {
	var data MergeRequestData
	path := fmt.Sprintf("projects/%d/merge_requests/%d", projectID, mergeRequestId)

	if err := doGitlabGet(path, &data); err != nil {
		return MergeRequestData{}, err
	}

	return MergeRequestData{
		data.Title,
		data.HasConflicts,
		data.Url,
		data.ChangesCount,
		data.SourceBranch,
		data.Description,
		commitHash,
	}, nil

}

type MergeRequestCommit struct {
	ID string `json:"id"`
}

type MergeRequestStats struct {
	Additions int     `json:"additions"`
	Deletions int     `json:"deletions"`
	Extra     *string `json:"extra,omitempty"`
}

type MergeRequestDiff struct {
	Diff    string `json:"diff"`
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	NewFile bool   `json:"new_file"`
}

type MergeRequestDiffResponse struct {
	Changes []MergeRequestDiff `json:"changes"`
}

func GetMergeRequestDiffs(projectID int, mergeRequestId int) ([]MergeRequestDiff, error) {
	var data MergeRequestDiffResponse
	path := fmt.Sprintf("projects/%d/merge_requests/%d/changes?access_raw_diffs=true", projectID, mergeRequestId)

	err := doGitlabGet(path, &data, false)
	if err != nil {
		return []MergeRequestDiff{}, err
	}

	return data.Changes, nil
}

func GetRawFile(projectID int, filePath, gitBranch string) (string, error) {
	var data string
	path := fmt.Sprintf("projects/%d/repository/files/%s/raw?ref=%s", projectID,
		url.QueryEscape(filePath),
		gitBranch)

	if err := doGitlabGet(path, &data); err != nil {
		return "", err
	}

	return data, nil
}

func GetCommitDiffs(projectID int, commitSHA string) ([]MergeRequestDiff, error) {
	var data []MergeRequestDiff
	path := fmt.Sprintf("projects/%d/repository/commits/%s/diff", projectID, commitSHA)

	err := doGitlabGet(path, &data, false)
	if err != nil {
		return []MergeRequestDiff{}, err
	}

	return data, nil
}
