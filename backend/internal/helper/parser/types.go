package parser

// DependencyInfo represents parsed dependency information
type DependencyInfo struct {
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	Repo         string `json:"repo"`
	Version      string `json:"version"`
	Runtime      string `json:"runtime"`
	GitHubURL    string `json:"github_url,omitempty"`
	IsGitHubRepo bool   `json:"is_github_repo"`
}

// GitHubRepoInfo contains verified GitHub repository information
type GitHubRepoInfo struct {
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	URL         string `json:"url"`
	Exists      bool   `json:"exists"`
	Description string `json:"description,omitempty"`
	Language    string `json:"language,omitempty"`
}

// ParseResult contains the parsing results
type ParseResult struct {
	Dependencies []DependencyInfo `json:"dependencies"`
	Runtime      string           `json:"runtime"`
	Success      bool             `json:"success"`
	Error        string           `json:"error,omitempty"`
}

// RuntimeType represents supported runtime types
type RuntimeType string

const (
	RuntimeGo      RuntimeType = "go"
	RuntimeNode    RuntimeType = "node"
	RuntimePython  RuntimeType = "python"
	RuntimeJava    RuntimeType = "java"
	RuntimeGradle  RuntimeType = "gradle"
	RuntimeDotNet  RuntimeType = "dotnet"
	RuntimeRuby    RuntimeType = "ruby"
	RuntimePHP     RuntimeType = "php"
	RuntimeRust    RuntimeType = "rust"
	RuntimeUnknown RuntimeType = "unknown"
)

// RuntimeParser interface that all specific parsers must implement
type RuntimeParser interface {
	Parse(content string) ([]DependencyInfo, error)
	GetRuntime() RuntimeType
	ParseDependency(name, version string) *DependencyInfo
}

// GitHubAPIInterface defines methods needed for GitHub repository verification
type GitHubAPIInterface interface {
	GetDefaultBranch(owner, repo string) (string, error)
	GetRepositoryInfo(owner, repo string) (map[string]interface{}, error)
}
