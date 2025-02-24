package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitlabUrlNotParsed(t *testing.T) {

	os.Setenv("GITLAB_DOMAIN", "https://gitlab.zxz.su/")

	expectedErrorMsg := "Переданная сссылка не содержит вадидного URL merge request"

	projectName, mergeRequestId, commitHash, err := ParseGitlabURL("")
	assert.Equal(t, "", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 0, mergeRequestId)
	assert.NotNil(t, err)
	assert.EqualError(t, err, expectedErrorMsg)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL("aaaaa")
	assert.Equal(t, "", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 0, mergeRequestId)
	assert.NotNil(t, err)
	assert.EqualError(t, err, expectedErrorMsg)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL("https://www.google.com/")
	assert.Equal(t, "", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 0, mergeRequestId)
	assert.NotNil(t, err)
	assert.EqualError(t, err, expectedErrorMsg)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/", os.Getenv("GITLAB_DOMAIN")))
	assert.Equal(t, "", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 0, mergeRequestId)
	assert.NotNil(t, err)
	assert.EqualError(t, err, expectedErrorMsg)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/kalbacker/-/merge_requests", os.Getenv("GITLAB_DOMAIN")))
	assert.Equal(t, "", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 0, mergeRequestId)
	assert.NotNil(t, err)
	assert.EqualError(t, err, expectedErrorMsg)
}

func TestParseGitlabUrlParsedWithSubDirectories(t *testing.T) {
	os.Setenv("GITLAB_DOMAIN", "https://gitlab.zxz.su/")

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")

	projectName, mergeRequestId, commitHash, err := ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/web-client/-/merge_requests/2507", gitlabDomain))
	assert.Equal(t, "web-client", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 2507, mergeRequestId)
	assert.Nil(t, err)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/web-client/-/merge_requests/2507/commits", gitlabDomain))
	assert.Equal(t, "web-client", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 2507, mergeRequestId)
	assert.Nil(t, err)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/web-client/-/merge_requests/2507/diffs", gitlabDomain))
	assert.Equal(t, "web-client", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 2507, mergeRequestId)
	assert.Nil(t, err)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/web-client/-/merge_requests/2507/pipelines", gitlabDomain))
	assert.Equal(t, "web-client", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 2507, mergeRequestId)
	assert.Nil(t, err)
}

func TestParseGitlabUrlParsedDifferentProjects(t *testing.T) {
	os.Setenv("GITLAB_DOMAIN", "gitlab.zxz.su")

	gitlabDomain := os.Getenv("GITLAB_DOMAIN")

	projectName, mergeRequestId, commitHash, err := ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/ads-backend/-/merge_requests/238", gitlabDomain))
	assert.Equal(t, "ads-backend", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 238, mergeRequestId)
	assert.Nil(t, err)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/cm-frontend/-/merge_requests/634", gitlabDomain))
	assert.Equal(t, "cm-frontend", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 634, mergeRequestId)
	assert.Nil(t, err)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc/arm/-/merge_requests/1401", gitlabDomain))
	assert.Equal(t, "arm", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 1401, mergeRequestId)
	assert.Nil(t, err)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc-admins/deploy/ugc/-/merge_requests/1760", gitlabDomain))
	assert.Equal(t, "ugc", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 1760, mergeRequestId)
	assert.Nil(t, err)

	projectName, mergeRequestId, commitHash, err = ParseGitlabURL(fmt.Sprintf("https://%s/ugc/web/arm-vector/-/merge_requests/242", gitlabDomain))
	assert.Equal(t, "arm-vector", projectName)
	assert.Equal(t, "", commitHash)
	assert.Equal(t, 242, mergeRequestId)
	assert.Nil(t, err)
}
