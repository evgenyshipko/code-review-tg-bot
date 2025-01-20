package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
	SourceBranch string `json:"source_branch"`
}

type CustomHTTPClient struct {
	http.Client
}

func (c *CustomHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("REQUEST %s %s FAILED WITH STATUS %d: %s", resp.Request.Method, resp.Request.URL, resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func GetProjectId(projectName string) (int, error) {

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	url := fmt.Sprintf("https://%s/api/v4/projects?search=%s&order_by=similarity", gitlabDomain, projectName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("не удалось создать реквест: %s", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &CustomHTTPClient{}
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

	client := &CustomHTTPClient{}
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
	Diff    string `json:"diff"`
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	NewFile bool   `json:"new_file"`
}

type MergeRequestDiffResponse []MergeRequestDiff

func GetMergeRequestDiffs(projectID int, mergeRequestId int, pageSize int) (MergeRequestDiffResponse, error) {

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	url1 := fmt.Sprintf("https://%s/api/v4/projects/%d/merge_requests/%d/diffs?per_page=%d", gitlabDomain, projectID, mergeRequestId, pageSize)

	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		return MergeRequestDiffResponse{}, fmt.Errorf("не удалось создать реквест: %s", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &CustomHTTPClient{}
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

func GetRawFile(projectID int, filePath string, gitBranch string) (string, error) {

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	url1 := fmt.Sprintf("https://%s/api/v4/projects/%d/repository/files/%s/raw?ref=%s", gitlabDomain, projectID, url.QueryEscape(filePath), gitBranch)

	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		return "", fmt.Errorf("не удалось создать реквест: %s", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &CustomHTTPClient{}
	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("не удалось выполнить запрос: %s", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	return string(body), nil
}
