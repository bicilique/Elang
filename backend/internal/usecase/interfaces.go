package usecase

import (
	"context"
	"elang-backend/internal/model"
)

// MessagingInterface defines methods for sending messages/notifications
type MessagingInterface interface {
	SendSummaryToTelegram(request model.TelegramSummary) error
}

/**
 * GitHubAPIInterface defines the methods for interacting with the GitHub API.
 * This interfaces using GraphQL and REST API to fetch repository information, commits, and file contents.
 */
type GitHubAPIInterface interface {
	GetDefaultBranch(owner, repo string) (string, error)
	GetListCommits(owner, repo, branch string) ([]map[string]interface{}, error)
	GetCommitsDetail(owner, repo, sha string) (*model.CommitDetail, error)
	GetFileContent(owner, repo, path, ref string) (string, error)
	GetRepoInfo(owner, repo string) (map[string]interface{}, error)
	ListBranches(owner, repo string) ([]string, error)
	ListTags(owner, repo string) ([]map[string]interface{}, error)
	ListPullRequests(owner, repo string, state string) ([]map[string]interface{}, error)
	GetPullRequestDetail(owner, repo string, number int) (map[string]interface{}, error)
	ListIssues(owner, repo string, state string) ([]map[string]interface{}, error)
	GetIssueDetail(owner, repo string, number int) (map[string]interface{}, error)
	ListDirectoryContents(owner, repo, path, ref string) ([]map[string]interface{}, error)
	GetUserInfo(username string) (map[string]interface{}, error)
	ListCollaborators(owner, repo string) ([]map[string]interface{}, error)
	ListWebhooks(owner, repo string) ([]map[string]interface{}, error)
	CompareCommits(owner, repo, base, head string) (*model.CompareCommitResult, error)
	FindMatchingTag(owner, repo, version string) (string, error)
}

// ObjectStorageInterface defines methods for object storage operations
type ObjectStorageInterface interface {
	// Analysis results
	// SaveAnalysisResult(ctx context.Context, result *helper.SecurityAnalysisResult) error
	// SaveAIResult(ctx context.Context, result *helper.SecurityAnalysisResult, AIAnalysis interface{}) error
	// GetAnalysisResult(ctx context.Context, objectKey string) (*helper.SecurityAnalysisResult, error)
	// GetAnalysisResultsByTimeRange(ctx context.Context, repository string, startTime, endTime time.Time) ([]*helper.SecurityAnalysisResult, error)
	// ListAnalysisResults(ctx context.Context, repository string) ([]string, error)
	// SearchAnalysisResults(ctx context.Context, filters map[string]string) ([]string, error)
	// DeleteAnalysisResult(ctx context.Context, objectKey string) error

	// SBOM operations
	SaveSBOM(ctx context.Context, appID string, appName string, sbomData []byte, format string) (string, error)
	GetSBOM(ctx context.Context, objectKey string) ([]byte, error)
	ListSBOMs(ctx context.Context, appName string) ([]string, error)

	// Vulnerability report operations
	SaveVulnerabilityReport(ctx context.Context, appID string, appName string, reportData []byte, format string) (string, error)
	GetVulnerabilityReport(ctx context.Context, objectKey string) ([]byte, error)
	ListVulnerabilityReports(ctx context.Context, appName string) ([]string, error)
}
