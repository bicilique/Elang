package helper

import (
	"elang-backend/internal/helper/parser"
	"fmt"
	"path/filepath"

	"strings"
)

// Type aliases for backward compatibility
type DependencyInfo = parser.DependencyInfo
type GitHubRepoInfo = parser.GitHubRepoInfo
type ParseResult = parser.ParseResult
type RuntimeType = parser.RuntimeType

// Runtime constants for backward compatibility
const (
	RuntimeGo      = parser.RuntimeGo
	RuntimeNode    = parser.RuntimeNode
	RuntimePython  = parser.RuntimePython
	RuntimeJava    = parser.RuntimeJava
	RuntimeGradle  = parser.RuntimeGradle
	RuntimeDotNet  = parser.RuntimeDotNet
	RuntimeRuby    = parser.RuntimeRuby
	RuntimePHP     = parser.RuntimePHP
	RuntimeRust    = parser.RuntimeRust
	RuntimeUnknown = parser.RuntimeUnknown
)

// GitHubAPIInterface defines methods needed for GitHub repository verification
type GitHubAPIInterface = parser.GitHubAPIInterface

// DependencyParser handles parsing of dependency files across different runtimes
type DependencyParser struct {
	parsers   map[parser.RuntimeType]parser.RuntimeParser
	githubAPI parser.GitHubAPIInterface // Optional: for repository verification
}

// NewDependencyParser creates a new instance of DependencyParser
func NewDependencyParser() *DependencyParser {
	dp := &DependencyParser{
		parsers: make(map[parser.RuntimeType]parser.RuntimeParser),
	}

	// Register parsers for different runtimes
	dp.parsers[parser.RuntimeGo] = parser.NewGoParser()
	dp.parsers[parser.RuntimeNode] = parser.NewNodeParser()
	dp.parsers[parser.RuntimePython] = parser.NewPythonParser()
	dp.parsers[parser.RuntimeJava] = parser.NewJavaParser()
	dp.parsers[parser.RuntimeGradle] = parser.NewGradleParser()
	dp.parsers[parser.RuntimeDotNet] = parser.NewDotNetParser()
	dp.parsers[parser.RuntimeRuby] = parser.NewRubyParser()
	dp.parsers[parser.RuntimePHP] = parser.NewPHPParser()
	dp.parsers[parser.RuntimeRust] = parser.NewRustParser()

	return dp
}

// NewDependencyParserWithGitHub creates a parser with GitHub API integration
func NewDependencyParserWithGitHub(githubAPI parser.GitHubAPIInterface) *DependencyParser {
	dp := NewDependencyParser()
	dp.githubAPI = githubAPI
	return dp
}

// DetectRuntime detects the runtime based on file content and filename
func (dp *DependencyParser) DetectRuntime(filename, content string) parser.RuntimeType {
	filename = strings.ToLower(filepath.Base(filename))

	switch filename {
	case "go.mod", "go.sum":
		return parser.RuntimeGo
	case "package.json", "package-lock.json", "yarn.lock":
		return parser.RuntimeNode
	case "requirements.txt", "pyproject.toml", "poetry.lock", "pipfile", "pipfile.lock":
		return parser.RuntimePython
	case "pom.xml":
		return parser.RuntimeJava
	case "build.gradle", "build.gradle.kts":
		return parser.RuntimeGradle
	case "gemfile", "gemfile.lock":
		return parser.RuntimeRuby
	case "composer.json", "composer.lock":
		return parser.RuntimePHP
	case "cargo.toml", "cargo.lock":
		return parser.RuntimeRust
	}

	// Check for .csproj, .vbproj, .fsproj extensions
	if strings.HasSuffix(filename, ".csproj") ||
		strings.HasSuffix(filename, ".vbproj") ||
		strings.HasSuffix(filename, ".fsproj") ||
		filename == "packages.config" {
		return parser.RuntimeDotNet
	}

	// Content-based detection as fallback
	if strings.Contains(content, "module ") && strings.Contains(content, "require") {
		return parser.RuntimeGo
	}
	if strings.Contains(content, "\"dependencies\"") && strings.Contains(content, "\"name\"") {
		return parser.RuntimeNode
	}
	if strings.Contains(content, "<project") && strings.Contains(content, "<dependencies>") {
		return parser.RuntimeJava
	}

	return parser.RuntimeUnknown
}

// ParseDependencyFile parses a dependency file and returns dependency information
func (dp *DependencyParser) ParseDependencyFile(filename, content string, runtimeHint ...parser.RuntimeType) parser.ParseResult {
	var runtime parser.RuntimeType

	// Use runtime hint if provided, otherwise detect
	if len(runtimeHint) > 0 && runtimeHint[0] != parser.RuntimeUnknown {
		runtime = runtimeHint[0]
	} else {
		runtime = dp.DetectRuntime(filename, content)
	}

	if runtime == parser.RuntimeUnknown {
		return parser.ParseResult{
			Success: false,
			Error:   "unable to detect runtime from file",
			Runtime: string(runtime),
		}
	}

	runtimeParser, exists := dp.parsers[runtime]
	if !exists {
		return parser.ParseResult{
			Success: false,
			Error:   fmt.Sprintf("no parser available for runtime: %s", runtime),
			Runtime: string(runtime),
		}
	}

	dependencies, err := runtimeParser.Parse(content)
	if err != nil {
		return parser.ParseResult{
			Success: false,
			Error:   err.Error(),
			Runtime: string(runtime),
		}
	}

	return parser.ParseResult{
		Dependencies: dependencies,
		Runtime:      string(runtime),
		Success:      true,
	}
}

// ParseDependencyFileWithGitHub parses a dependency file and verifies GitHub repositories
func (dp *DependencyParser) ParseDependencyFileWithGitHub(filename, content string, runtimeHint ...parser.RuntimeType) parser.ParseResult {
	result := dp.ParseDependencyFile(filename, content, runtimeHint...)

	if !result.Success {
		return result
	}

	// Enhance dependencies with GitHub repository information
	for i := range result.Dependencies {
		dp.enhanceWithGitHubInfo(&result.Dependencies[i])
	}

	return result
}

// enhanceWithGitHubInfo adds GitHub repository information to a dependency
func (dp *DependencyParser) enhanceWithGitHubInfo(dep *parser.DependencyInfo) {
	// First, try to construct GitHub URL from known patterns
	githubURL := dp.constructGitHubURL(dep)
	dep.GitHubURL = githubURL

	if githubURL != "" {
		dep.IsGitHubRepo = true

		// If we have GitHub API access, verify the repository exists
		if dp.githubAPI != nil && dep.Owner != "" && dep.Repo != "" {
			_, err := dp.githubAPI.GetDefaultBranch(dep.Owner, dep.Repo)
			if err != nil {
				dep.IsGitHubRepo = false
				dep.GitHubURL = ""
			}
		}
	}
}

// constructGitHubURL attempts to construct a GitHub URL from dependency information
func (dp *DependencyParser) constructGitHubURL(dep *parser.DependencyInfo) string {
	// Get the specific parser for this runtime
	runtimeParser, exists := dp.parsers[parser.RuntimeType(dep.Runtime)]
	if !exists {
		// Default case: if we have owner/repo, assume GitHub
		if dep.Owner != "" && dep.Repo != "" {
			return fmt.Sprintf("https://github.com/%s/%s", dep.Owner, dep.Repo)
		}
		return ""
	}

	// Try to get URL using runtime-specific logic
	switch p := runtimeParser.(type) {
	case *parser.GoParser:
		return p.GetRepositoryURL(dep)
	case *parser.NodeParser:
		return p.GetRepositoryURL(dep)
	case *parser.PythonParser:
		return p.GetRepositoryURL(dep)
	case *parser.JavaParser:
		return p.GetRepositoryURL(dep)
	case *parser.GradleParser:
		return p.GetRepositoryURL(dep)
	case *parser.DotNetParser:
		return p.GetRepositoryURL(dep)
	case *parser.RubyParser:
		return p.GetRepositoryURL(dep)
	case *parser.PHPParser:
		return p.GetRepositoryURL(dep)
	case *parser.RustParser:
		return p.GetRepositoryURL(dep)
	default:
		// Default case: if we have owner/repo, assume GitHub
		if dep.Owner != "" && dep.Repo != "" {
			return fmt.Sprintf("https://github.com/%s/%s", dep.Owner, dep.Repo)
		}
	}

	return ""
}

// VerifyGitHubRepository verifies if a GitHub repository exists and gets additional info
func (dp *DependencyParser) VerifyGitHubRepository(owner, repo string) (*parser.GitHubRepoInfo, error) {
	if dp.githubAPI == nil {
		return nil, fmt.Errorf("GitHub API not configured")
	}

	// Try to get repository information
	repoInfo, err := dp.githubAPI.GetRepositoryInfo(owner, repo)
	if err != nil {
		return &parser.GitHubRepoInfo{
			Owner:  owner,
			Repo:   repo,
			URL:    fmt.Sprintf("https://github.com/%s/%s", owner, repo),
			Exists: false,
		}, err
	}

	info := &parser.GitHubRepoInfo{
		Owner:  owner,
		Repo:   repo,
		URL:    fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		Exists: true,
	}

	// Extract additional information if available
	if desc, ok := repoInfo["description"].(string); ok {
		info.Description = desc
	}
	if lang, ok := repoInfo["language"].(string); ok {
		info.Language = lang
	}

	return info, nil
}

// RuntimeNameToType maps human-readable runtime names to internal RuntimeType constants
var RuntimeNameToType = map[string]parser.RuntimeType{
	"Go":      parser.RuntimeGo,
	"Node.js": parser.RuntimeNode,
	"Python":  parser.RuntimePython,
	"Java":    parser.RuntimeJava,
	"Gradle":  parser.RuntimeGradle,
	"DotNet":  parser.RuntimeDotNet,
	"Ruby":    parser.RuntimeRuby,
	"PHP":     parser.RuntimePHP,
	"Rust":    parser.RuntimeRust,
}

// RuntimeTypeToName maps internal RuntimeType constants to human-readable names
var RuntimeTypeToName = map[parser.RuntimeType]string{
	parser.RuntimeGo:      "Go",
	parser.RuntimeNode:    "Node.js",
	parser.RuntimePython:  "Python",
	parser.RuntimeJava:    "Java",
	parser.RuntimeGradle:  "Gradle",
	parser.RuntimeDotNet:  "DotNet",
	parser.RuntimeRuby:    "Ruby",
	parser.RuntimePHP:     "PHP",
	parser.RuntimeRust:    "Rust",
	parser.RuntimeUnknown: "Unknown",
}

// RuntimeNameToTypeCI maps lowercased runtime names to internal RuntimeType constants (case-insensitive)
var RuntimeNameToTypeCI = map[string]parser.RuntimeType{}

func init() {
	for k, v := range RuntimeNameToType {
		RuntimeNameToTypeCI[strings.ToLower(k)] = v
	}
}

// GetRuntimeTypeCI returns the RuntimeType for a given name, case-insensitive
func GetRuntimeTypeCI(name string) parser.RuntimeType {
	if rt, ok := RuntimeNameToTypeCI[strings.ToLower(name)]; ok {
		return rt
	}
	return parser.RuntimeUnknown
}
