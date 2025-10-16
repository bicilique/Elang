package parser

import (
	"bufio"
	"regexp"
	"strings"
)

// GoParser handles parsing of Go module files
type GoParser struct{}

// NewGoParser creates a new instance of GoParser
func NewGoParser() *GoParser {
	return &GoParser{}
}

// GetRuntime returns the runtime type for Go
func (p *GoParser) GetRuntime() RuntimeType {
	return RuntimeGo
}

// Parse parses go.mod files
func (p *GoParser) Parse(content string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo

	// Regex patterns for go.mod parsing
	requireBlockRegex := regexp.MustCompile(`require\s*\(\s*([\s\S]*?)\s*\)`)
	requireLineRegex := regexp.MustCompile(`require\s+([^\s]+)\s+([^\s]+)`)
	dependencyRegex := regexp.MustCompile(`^\s*([^\s]+)\s+([^\s/]+(?:\s+//\s*indirect)?)`)

	// Handle require blocks
	requireBlocks := requireBlockRegex.FindAllStringSubmatch(content, -1)
	for _, block := range requireBlocks {
		blockContent := block[1]
		scanner := bufio.NewScanner(strings.NewReader(blockContent))

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}

			// Skip indirect dependencies
			if strings.Contains(line, "// indirect") {
				continue
			}

			// Remove inline comments
			if idx := strings.Index(line, "//"); idx != -1 {
				line = strings.TrimSpace(line[:idx])
			}

			matches := dependencyRegex.FindStringSubmatch(line)
			if len(matches) >= 3 {
				depInfo := p.ParseDependency(matches[1], matches[2])
				if depInfo != nil {
					dependencies = append(dependencies, *depInfo)
				}
			}
		}
	}

	// Handle single require lines
	requireLines := requireLineRegex.FindAllStringSubmatch(content, -1)
	for _, match := range requireLines {
		if len(match) >= 3 && !strings.Contains(match[2], "indirect") {
			depInfo := p.ParseDependency(match[1], match[2])
			if depInfo != nil {
				dependencies = append(dependencies, *depInfo)
			}
		}
	}

	return dependencies, nil
}

// ParseDependency parses a single Go dependency
func (p *GoParser) ParseDependency(name, version string) *DependencyInfo {
	// Handle indirect dependencies - skip them
	if strings.Contains(version, "indirect") {
		return nil
	}

	// Extract owner and repo from module path
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		return nil
	}

	var owner, repo string

	// Handle common patterns
	if strings.Contains(name, "github.com") && len(parts) >= 3 {
		owner = parts[1]
		repo = parts[2]
	} else if strings.Contains(name, "gitlab.com") && len(parts) >= 3 {
		owner = parts[1]
		repo = parts[2]
	} else if strings.Contains(name, "bitbucket.org") && len(parts) >= 3 {
		owner = parts[1]
		repo = parts[2]
	} else if strings.HasPrefix(name, "gorm.io/") && len(parts) >= 2 {
		// Special handling for gorm.io packages
		if parts[1] == "driver" && len(parts) >= 3 {
			owner = "go-gorm"
			repo = parts[2]
		} else {
			owner = "go-gorm"
			repo = parts[len(parts)-1]
		}
	} else {
		// For other cases, try to extract meaningful parts
		if len(parts) >= 2 {
			owner = parts[len(parts)-2]
			repo = parts[len(parts)-1]
		}
	}

	return &DependencyInfo{
		Name:    name,
		Owner:   owner,
		Repo:    repo,
		Version: version,
		Runtime: string(RuntimeGo),
	}
}

// GetRepositoryURL gets GitHub URL for Go modules
func (p *GoParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Handle special cases for popular Go libraries
	specialCases := map[string]string{
		"gorm.io/gorm":            "https://github.com/go-gorm/gorm",
		"gorm.io/driver/postgres": "https://github.com/go-gorm/postgres",
		"gorm.io/driver/mysql":    "https://github.com/go-gorm/mysql",
		"gorm.io/driver/sqlite":   "https://github.com/go-gorm/sqlite",
		"google.golang.org/grpc":  "https://github.com/grpc/grpc-go",
		"golang.org/x/crypto":     "https://github.com/golang/crypto",
		"golang.org/x/net":        "https://github.com/golang/net",
		"golang.org/x/sys":        "https://github.com/golang/sys",
		"golang.org/x/text":       "https://github.com/golang/text",
		"gopkg.in/yaml.v2":        "https://github.com/go-yaml/yaml",
		"gopkg.in/yaml.v3":        "https://github.com/go-yaml/yaml",
	}

	if url, exists := specialCases[dep.Name]; exists {
		return url
	}

	// Default GitHub pattern
	if dep.Owner != "" && dep.Repo != "" {
		return "https://github.com/" + dep.Owner + "/" + dep.Repo
	}

	return ""
}
