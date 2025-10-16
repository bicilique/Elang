package usecase

import (
	"bytes"
	"elang-backend/internal/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type GithubAPIusecase struct {
	// Add necessary fields, e.g., HTTP client, authentication tokens, etc.
	Token      string
	HTTPClient *http.Client
}

func NewGitHubAPIusecase(token string) GitHubAPIInterface {
	return &GithubAPIusecase{
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

// GetDefaultBranch fetches the default branch of a given repository.
// Uses REST API if no token is provided, otherwise uses GraphQL API.
func (g *GithubAPIusecase) GetDefaultBranch(owner, repo string) (string, error) {
	// If no token, use REST API instead of GraphQL
	if g.Token == "" {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
		log.Println("Request URL:", url)
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", err
		}
		request.Header.Set("Accept", "application/vnd.github.v3+json")
		resp, err := g.HTTPClient.Do(request)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		log.Println("Response Status:", resp.Status)
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("GitHub API returned status: %s", resp.Status)
		}
		var result struct {
			DefaultBranch string `json:"default_branch"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
		return result.DefaultBranch, nil
	}

	// Use GraphQL API when token is available
	query := fmt.Sprintf(`query { repository(owner: "%s", name: "%s") { defaultBranchRef { name } } }`, owner, repo)
	resp, err := g.doGraphQLRequest(query)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub GraphQL API returned status: %s", resp.Status)
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
	return result.Data.Repository.DefaultBranchRef.Name, nil
}

// GetListCommits fetches the list of commits for a given branch.
// Uses REST API if no token is provided, otherwise uses GraphQL API.
func (g *GithubAPIusecase) GetListCommits(owner, repo, branch string) ([]map[string]interface{}, error) {
	// If no token, use REST API instead of GraphQL
	if g.Token == "" {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?sha=%s&per_page=10", owner, repo, branch)
		log.Println("Request URL:", url)
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Accept", "application/vnd.github.v3+json")
		resp, err := g.HTTPClient.Do(request)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		log.Println("Response Status:", resp.Status)
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
		}
		var rawCommits []struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Name  string `json:"name"`
					Email string `json:"email"`
					Date  string `json:"date"`
				} `json:"author"`
				Committer struct {
					Name  string `json:"name"`
					Email string `json:"email"`
					Date  string `json:"date"`
				} `json:"committer"`
			} `json:"commit"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&rawCommits); err != nil {
			return nil, err
		}
		var commits []map[string]interface{}
		for _, rc := range rawCommits {
			commit := map[string]interface{}{
				"oid":             rc.SHA,
				"message":         rc.Commit.Message,
				"author_name":     rc.Commit.Author.Name,
				"author_email":    rc.Commit.Author.Email,
				"author_date":     rc.Commit.Author.Date,
				"committer_name":  rc.Commit.Committer.Name,
				"committer_email": rc.Commit.Committer.Email,
				"committer_date":  rc.Commit.Committer.Date,
				"changed_files":   0, // REST API doesn't provide this in list view
			}
			commits = append(commits, commit)
		}
		return commits, nil
	}

	// Use GraphQL API when token is available
	query := fmt.Sprintf(`query { repository(owner: "%s", name: "%s") { ref(qualifiedName: "%s") { target { ... on Commit { history(first: 10) { edges { node { oid message author { name email date } committer { name email date } changedFiles } } } } } } } }`, owner, repo, branch)
	resp, err := g.doGraphQLRequest(query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub GraphQL API returned status: %s", resp.Status)
	}
	var result struct {
		Data struct {
			Repository struct {
				Ref struct {
					Target struct {
						History struct {
							Edges []struct {
								Node struct {
									Oid     string `json:"oid"`
									Message string `json:"message"`
									Author  struct {
										Name  string `json:"name"`
										Email string `json:"email"`
										Date  string `json:"date"`
									} `json:"author"`
									Committer struct {
										Name  string `json:"name"`
										Email string `json:"email"`
										Date  string `json:"date"`
									} `json:"committer"`
									ChangedFiles int `json:"changedFiles"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"history"`
					} `json:"target"`
				} `json:"ref"`
			} `json:"repository"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var commits []map[string]interface{}
	for _, edge := range result.Data.Repository.Ref.Target.History.Edges {
		commit := map[string]interface{}{
			"oid":             edge.Node.Oid,
			"message":         edge.Node.Message,
			"author_name":     edge.Node.Author.Name,
			"author_email":    edge.Node.Author.Email,
			"author_date":     edge.Node.Author.Date,
			"committer_name":  edge.Node.Committer.Name,
			"committer_email": edge.Node.Committer.Email,
			"committer_date":  edge.Node.Committer.Date,
			"changed_files":   edge.Node.ChangedFiles,
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

// GetCommitsDetail fetches commit details using the GitHub REST API for a given commit SHA.
func (g *GithubAPIusecase) GetCommitsDetail(owner, repo, sha string) (*model.CommitDetail, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, sha)
	log.Println("Request URL:", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	// Unmarshal into a temporary struct to extract nested fields
	var raw struct {
		SHA    string `json:"sha"`
		Commit struct {
			Author    model.CommitPerson `json:"author"`
			Committer model.CommitPerson `json:"committer"`
			Message   string             `json:"message"`
		} `json:"commit"`
		Stats model.CommitStats  `json:"stats"`
		Files []model.CommitFile `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	result := &model.CommitDetail{
		SHA:       raw.SHA,
		Author:    raw.Commit.Author,
		Committer: raw.Commit.Committer,
		Message:   raw.Commit.Message,
		Stats:     raw.Stats,
		Files:     raw.Files,
	}
	return result, nil
}

// GetFileContent fetches the raw content of a file at a specific commit (ref) using the GitHub REST API.
func (g *GithubAPIusecase) GetFileContent(owner, repo, path, ref string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, ref)
	log.Println("Request URL:", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3.raw") // Get raw file content
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	contentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}

// GetRepoInfo fetches repository information using the GitHub REST API.
func (g *GithubAPIusecase) GetRepoInfo(owner, repo string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	log.Println("Request URL:", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListBranches lists all branches in a repository.
func (g *GithubAPIusecase) ListBranches(owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repo)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var branches []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}
	var names []string
	for _, b := range branches {
		names = append(names, b.Name)
	}
	return names, nil
}

// ListTags lists all tags in a repository.
func (g *GithubAPIusecase) ListTags(owner, repo string) ([]map[string]interface{}, error) {
	var defaultNumberOfTags = "100"
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags?per_page=%s", owner, repo, defaultNumberOfTags)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, t := range tags {
		result = append(result, map[string]interface{}{
			"name":       t.Name,
			"commit_sha": t.Commit.SHA,
		})
	}
	return result, nil
}

// ListPullRequests lists pull requests for a repository.
func (g *GithubAPIusecase) ListPullRequests(owner, repo, state string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=%s", owner, repo, state)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var pulls []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pulls); err != nil {
		return nil, err
	}
	return pulls, nil
}

// GetPullRequestDetail gets details of a specific pull request.
func (g *GithubAPIusecase) GetPullRequestDetail(owner, repo string, number int) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var pr map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return pr, nil
}

// ListIssues lists issues for a repository.
func (g *GithubAPIusecase) ListIssues(owner, repo, state string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=%s", owner, repo, state)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var issues []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// GetIssueDetail gets details of a specific issue.
func (g *GithubAPIusecase) GetIssueDetail(owner, repo string, number int) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, number)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var issue map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}
	return issue, nil
}

// ListDirectoryContents lists files and directories at a given path/ref.
func (g *GithubAPIusecase) ListDirectoryContents(owner, repo, path, ref string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, ref)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var contents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}
	return contents, nil
}

// GetUserInfo gets information about a GitHub user.
func (g *GithubAPIusecase) GetUserInfo(username string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return user, nil
}

// ListCollaborators lists collaborators for a repository.
func (g *GithubAPIusecase) ListCollaborators(owner, repo string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/collaborators", owner, repo)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

// ListWebhooks lists webhooks configured for a repository.
func (g *GithubAPIusecase) ListWebhooks(owner, repo string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var hooks []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&hooks); err != nil {
		return nil, err
	}
	return hooks, nil
}

// CompareCommits compares two commits (base and head) in a repository using GitHub's REST API.
func (g *GithubAPIusecase) CompareCommits(owner, repo, base, head string) (*model.CompareCommitResult, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/compare/%s...%s", owner, repo, base, head)
	log.Println("CompareCommits request URL:", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.Token != "" {
		request.Header.Set("Authorization", "token "+g.Token)
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}
	var result model.CompareCommitResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FindMatchingTag returns the tag name that matches or is most similar to the given version string
func (g *GithubAPIusecase) FindMatchingTag(owner, repo, version string) (string, error) {
	tags, err := g.ListTags(owner, repo)
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", nil
	}

	// Normalize version: add REL and rel to prefixes
	versionNorm := strings.ToLower(version)
	prefixes := []string{"v", "release-", "tags/", "tag/", "rel", "REL", "r"}
	for _, p := range prefixes {
		versionNorm = strings.TrimPrefix(versionNorm, strings.ToLower(p))
	}
	versionNorm = strings.TrimSpace(versionNorm)

	// Exact match first
	for _, t := range tags {
		if name, ok := t["name"].(string); ok && strings.EqualFold(name, version) {
			return name, nil
		}
	}

	// Fuzzy match: normalize tag name and compare
	for _, t := range tags {
		if name, ok := t["name"].(string); ok {
			nameNorm := strings.ToLower(name)
			for _, p := range prefixes {
				nameNorm = strings.TrimPrefix(nameNorm, strings.ToLower(p))
			}
			nameNorm = strings.TrimSpace(nameNorm)
			if nameNorm == versionNorm || strings.Contains(nameNorm, versionNorm) || strings.Contains(versionNorm, nameNorm) {
				return name, nil
			}
		}
	}

	return "", nil
}

// doGraphQLRequest is a reusable helper for sending GraphQL queries to GitHub
func (g *GithubAPIusecase) doGraphQLRequest(query string) (*http.Response, error) {
	graphqlURL := "https://api.github.com/graphql"
	body := map[string]interface{}{
		"query": query,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	log.Println("GraphQL Request URL:", graphqlURL)
	log.Printf("GraphQL Query: %s\n", query)
	request, err := http.NewRequest("POST", graphqlURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if g.Token != "" {
		request.Header.Set("Authorization", "bearer "+g.Token)
	}
	resp, err := g.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
