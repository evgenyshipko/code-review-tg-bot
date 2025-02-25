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
//   - commitHash - хэш коммита, если передана ссылка на коммит
func ParseGitlabURL(url1 string) (encodedQueryPathProject string, mergeRequestId int, commitHash string, err error) {
	gitlabDomain := os.Getenv("GITLAB_DOMAIN")

	regex := fmt.Sprintf(`https:\/\/%s\/([a-zA-Z-/]+)\/([a-zA-Z-]+)\/-\/merge_requests\/(\d+)(?:\/diffs\?commit_id=([a-f0-9]+))?.*`, gitlabDomain)

	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(url1)

	errMsg := "переданная сссылка не содержит вадидного URL merge request"

	if len(match) == 0 {
		return "", 0, "", errors.New(errMsg)
	}

	mergeRequestIdInteger, err := strconv.Atoi(match[3])
	if err != nil {
		return "", 0, "", errors.New(errMsg + err.Error())
	}

	encodedQueryPathProject = url.QueryEscape(match[1] + "/" + match[2])
	commitHash = match[4]

	if commitHash != "" {
		validLengthForCommitHash := 40
		if len(commitHash) != validLengthForCommitHash {
			errMsg = "переданная ссылка содержит некорректный хэш коммита"
			return "", 0, "", errors.New(errMsg)
		}
	}

	return encodedQueryPathProject, mergeRequestIdInteger, commitHash, nil
}
