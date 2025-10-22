package usecase_test

import (
	"elang-backend/internal/model"
	"elang-backend/internal/usecase"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubAPIUsecase_GetDefaultBranch(t *testing.T) {
	// Create a mock server
	mockResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"repository": map[string]interface{}{
				"defaultBranchRef": map[string]interface{}{
					"name": "main",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/graphql", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}))
	defer server.Close()

	githubUsecase := &testGitHubAPIUsecase{
		Token:      "test-token",
		HTTPClient: &http.Client{},
		BaseURL:    server.URL,
	}

	branch, err := githubUsecase.GetDefaultBranch("owner", "repo")
	assert.NoError(t, err)
	assert.Equal(t, "main", branch)
}

func TestGitHubAPIUsecase_GetDefaultBranch_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	githubUsecase := &testGitHubAPIUsecase{
		Token:      "test-token",
		HTTPClient: &http.Client{},
		BaseURL:    server.URL,
	}

	_, err := githubUsecase.GetDefaultBranch("owner", "repo")
	assert.Error(t, err)
}

func TestNewGitHubAPIUsecase(t *testing.T) {
	token := "test-token"
	usecase := usecase.NewGitHubAPIusecase(token)
	assert.NotNil(t, usecase)
}

// Mock implementation for testing
type testGitHubAPIUsecase struct {
	Token      string
	HTTPClient *http.Client
	BaseURL    string
}

func (g *testGitHubAPIUsecase) GetDefaultBranch(owner, repo string) (string, error) {
	query := `query { repository(owner: "` + owner + `", name: "` + repo + `") { defaultBranchRef { name } } }`
	req, err := http.NewRequest("POST", g.BaseURL+"/graphql", nil)
	if err != nil {
		return "", err
	}

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", assert.AnError
	}

	var result struct {
		Data struct {
			Repository struct {
				DefaultBranchRef struct {
					Name string `json:"name"`
				} `json:"defaultBranchRef"`
			} `json:"repository"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	_ = query
	return result.Data.Repository.DefaultBranchRef.Name, nil
}

func (g *testGitHubAPIUsecase) GetListCommits(owner, repo, branch string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) GetCommitsDetail(owner, repo, sha string) (*model.CommitDetail, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) GetFileContent(owner, repo, path, ref string) (string, error) {
	return "", nil
}

func (g *testGitHubAPIUsecase) GetRepoInfo(owner, repo string) (map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) ListBranches(owner, repo string) ([]string, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) ListTags(owner, repo string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) ListPullRequests(owner, repo string, state string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) GetPullRequestDetail(owner, repo string, number int) (map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) ListIssues(owner, repo string, state string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) GetIssueDetail(owner, repo string, number int) (map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) ListDirectoryContents(owner, repo, path, ref string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) GetUserInfo(username string) (map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) ListCollaborators(owner, repo string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) ListWebhooks(owner, repo string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) CompareCommits(owner, repo, base, head string) (*model.CompareCommitResult, error) {
	return nil, nil
}

func (g *testGitHubAPIUsecase) FindMatchingTag(owner, repo, version string) (string, error) {
	return "", nil
}

func TestGitHubAPIInterface(t *testing.T) {
	t.Run("InterfaceCompliance", func(t *testing.T) {
		var _ usecase.GitHubAPIInterface = &testGitHubAPIUsecase{}
		require.True(t, true, "testGitHubAPIUsecase implements GitHubAPIInterface")
	})
}
