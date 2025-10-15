package services

import (
	"context"
	"elang-backend/internal/entity"
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

type DependenciesInterface interface {
	// Scan Application for vulnerabilities by checking dependency versions in OSV
	ScanDependencies(ctx context.Context, appName, runtime, version, description, fileName, content string) (interface{}, error)

	// Get SBOM by its ID
	GetSBOMById(ctx context.Context, appName, sbomID string) ([]byte, error)

	// Start monitoring an application
	StartMonitoringApplication(ctx context.Context, appUID string) error

	// Stop monitoring an application
	StopMonitoringApplication(ctx context.Context, appUID string) error

	// Get monitoring status of an application
	GetMonitoringStatus(ctx context.Context, appUID string) (map[string]interface{}, error)
}

type DepedencyMonitoringInterface interface {
	// MonitorApplicationDepedencies starts monitoring an application's dependencies for changes
	MonitorApplicationDepedencies(ctx context.Context, app *entity.App) (interface{}, error)

	// StopMonitoringApplication stops monitoring an application's dependencies
	StopMonitoringApplication(ctx context.Context, app *entity.App) error

	// GetMonitoringStatus retrieves the monitoring status of an application
	GetMonitoringStatus(ctx context.Context, app *entity.App) (map[string]interface{}, error)
}

type ApplicationInterface interface {
	// Add or intialize Application -> input app name , depedency file , runtime type , description
	AddApplication(ctx context.Context, appName, runtimeType, framework, description, fileName, content string) (*model.AddApplicationResponse, error)

	// Add depedency to Application (batch)
	AddApplicationDependency(ctx context.Context, appUID string, deps []model.DependencyInfoRequest) (interface{}, error)

	// List Applications Dependency
	ListApplicationDependency(ctx context.Context, appUID string) (*model.ListApplicationDependencyResponse, error)

	// Update Application Dependency
	UpdateApplicationDependency(ctx context.Context, appUID string, input *model.UpdateApplicationDependencyRequest) (*model.UpdateApplicationDependencyResponse, error)

	// Remove depedency from Application (batch)
	RemoveApplicationDependency(ctx context.Context, appUID string, deps []string) (interface{}, error)

	// Remove Application or Deactivate Application
	RemoveApplication(ctx context.Context, appUID string) error

	// Recover Application or Reactivate Application
	RecoverApplication(ctx context.Context, appUID string) error

	// List Applications
	ListApplications(ctx context.Context) (*model.ListApplicationsResponse, error)

	// // Get Monitoring Status of Application
	GetApplicationStatus(ctx context.Context, appUID string) (map[string]interface{}, error)

	ScanApplicationDependencies(ctx context.Context, appUID string) (interface{}, error)

	// Get SBOM for an application
	GetApplicationSBOM(ctx context.Context, appUID string) ([]byte, error)

	// List all SBOMs for an application
	ListApplicationSBOMs(ctx context.Context, appUID string) ([]string, error)

	// // Get Monitoring Status of All Applications
	// GetAllApplicationsStatus(ctx context.Context) (map[string]interface{}, error)
}
