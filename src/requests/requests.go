package requests

import (
	"encoding/json"
	"fmt"
	"moul.io/http2curl"
	"net/http"
	"os"
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
}

func GetProjectId(projectName string) (int, error) {

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	url := fmt.Sprintf("https://%s/api/v4/projects?search=%s", gitlabDomain, projectName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("не удалось создать реквест: %s", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return 0, fmt.Errorf("не удалось выполнить запрос: %s", err)
	}

	defer resp.Body.Close()

	var data []ProjectData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return 0, fmt.Errorf("ошибка декода тела ответа: %s", err)
	}

	return data[0].ID, nil
}

func GetMergeRequestData(projectID int, mergeRequestId int) (MergeRequestData, error) {

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	url := fmt.Sprintf("https://%s/api/v4/projects/%d/merge_requests/%d", gitlabDomain, projectID, mergeRequestId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return MergeRequestData{}, fmt.Errorf("не удалось создать реквест: %s", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	command, _ := http2curl.GetCurlCommand(req)
	fmt.Println(command)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return MergeRequestData{}, fmt.Errorf("не удалось выполнить запрос: %s", err)
	}

	defer resp.Body.Close()

	var data MergeRequestData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return MergeRequestData{}, fmt.Errorf("ошибка декода тела ответа: %s", err)
	}

	return data, nil
}

type MergeRequestCommit struct {
	ID string `json:"id"`
}

type MergeRequestStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

type MergeRequestDiff struct {
	Diff string `json:"diff"`
}

type MergeRequestDiffResponse []MergeRequestDiff

func GetMergeRequestDiffs(projectID int, mergeRequestId int) (MergeRequestDiffResponse, error) {

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	url := fmt.Sprintf("https://%s/api/v4/projects/%d/merge_requests/%d/diffs", gitlabDomain, projectID, mergeRequestId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return MergeRequestDiffResponse{}, fmt.Errorf("не удалось создать реквест: %s", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return MergeRequestDiffResponse{}, fmt.Errorf("не удалось выполнить запрос: %s", err)
	}

	defer resp.Body.Close()

	var data MergeRequestDiffResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return MergeRequestDiffResponse{}, fmt.Errorf("ошибка декода тела ответа: %s", err)
	}

	return data, nil
}
