package parser

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func ParseGitlabURL(url string) (projectName string, mergeRequestId int, err error) {
	gitlabDomain := os.Getenv("GITLAB_DOMAIN")

	regex := fmt.Sprintf(`https:\/\/%s\/([a-zA-Z-/]+)\/([a-zA-Z-]+)\/-\/merge_requests\/(\d+).*`, gitlabDomain)

	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(url)

	errMsg := "Переданная сссылка не содержит вадидного URL merge request"

	if len(match) == 0 {
		return "", 0, errors.New(errMsg)
	}

	mergeRequestIdInteger, err := strconv.Atoi(match[3])
	if err != nil {
		return "", 0, errors.New(errMsg + err.Error())
	}

	projectName = match[2]
	mergeRequestId = mergeRequestIdInteger
	return
}
