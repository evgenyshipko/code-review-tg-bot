package requests

import (
	"encoding/json"
	"fmt"
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
