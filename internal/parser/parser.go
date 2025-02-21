package parser

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

// Парсит ссылу на МР и возвращает:
//   - encodedQueryPathProject - кодированный путь к проекту
//   - mergeRequestId - номер МР
func ParseGitlabURL(url1 string) (encodedQueryPathProject string, mergeRequestId int, err error) {
	gitlabDomain := os.Getenv("GITLAB_DOMAIN")

	regex := fmt.Sprintf(`https:\/\/%s\/([a-zA-Z-/]+)\/([a-zA-Z-]+)\/-\/merge_requests\/(\d+).*`, gitlabDomain)

	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(url1)

	errMsg := "Переданная сссылка не содержит вадидного URL merge request"

	if len(match) == 0 {
		return "", 0, errors.New(errMsg)
	}

	mergeRequestIdInteger, err := strconv.Atoi(match[3])
	if err != nil {
		return "", 0, errors.New(errMsg + err.Error())
	}

	encodedQueryPathProject = url.QueryEscape(match[1] + "/" + match[2])

	fmt.Printf("match: %v\n", match)
	mergeRequestId = mergeRequestIdInteger
	return
}
