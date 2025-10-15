package parser

import (
	"encoding/json"
	"regexp"
	"strings"
)

// NodeParser handles parsing of Node.js package files
type NodeParser struct{}

// NewNodeParser creates a new instance of NodeParser
func NewNodeParser() *NodeParser {
	return &NodeParser{}
}

// GetRuntime returns the runtime type for Node.js
func (p *NodeParser) GetRuntime() RuntimeType {
	return RuntimeNode
}

// Parse parses package.json files
func (p *NodeParser) Parse(content string) ([]DependencyInfo, error) {
	var packageJSON struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	// Handle empty or whitespace-only content
	content = strings.TrimSpace(content)
	if content == "" {
		return []DependencyInfo{}, nil
	}

	// Try to clean JSON with comments (although not standard JSON)
	if strings.Contains(content, "//") || strings.Contains(content, "/*") {
		// Remove line comments
		lines := strings.Split(content, "\n")
		var cleanedLines []string
		for _, line := range lines {
			if commentIdx := strings.Index(line, "//"); commentIdx != -1 {
				line = line[:commentIdx]
			}
			cleanedLines = append(cleanedLines, line)
		}
		content = strings.Join(cleanedLines, "\n")

		// Remove block comments (simple approach)
		content = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(content, "")
	}

	// If content is just comments or empty after cleaning, return empty result
	content = strings.TrimSpace(content)
	if content == "" || content == "{}" {
		return []DependencyInfo{}, nil
	}

	if err := json.Unmarshal([]byte(content), &packageJSON); err != nil {
		// If JSON parsing fails, return empty dependencies instead of error for edge cases
		return []DependencyInfo{}, nil
	}

	var dependencies []DependencyInfo

	// Parse regular dependencies
	for name, version := range packageJSON.Dependencies {
		if depInfo := p.ParseDependency(name, version); depInfo != nil {
			dependencies = append(dependencies, *depInfo)
		}
	}

	// Parse dev dependencies
	for name, version := range packageJSON.DevDependencies {
		if depInfo := p.ParseDependency(name, version); depInfo != nil {
			dependencies = append(dependencies, *depInfo)
		}
	}

	return dependencies, nil
}

// ParseDependency parses a single npm dependency
func (p *NodeParser) ParseDependency(name, version string) *DependencyInfo {
	// Handle scoped packages
	if strings.HasPrefix(name, "@") {
		parts := strings.Split(name[1:], "/")
		if len(parts) >= 2 {
			return &DependencyInfo{
				Name:    name,
				Owner:   parts[0],
				Repo:    parts[1],
				Version: version,
				Runtime: string(RuntimeNode),
			}
		}
	}

	// For regular packages, we can't easily determine owner/repo without API calls
	// But we'll provide the name and version
	return &DependencyInfo{
		Name:    name,
		Owner:   "", // Would need npm API to determine
		Repo:    name,
		Version: version,
		Runtime: string(RuntimeNode),
	}
}

// GetRepositoryURL gets GitHub URL for npm packages
func (p *NodeParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Common npm packages with known GitHub repositories
	commonPackages := map[string]string{
		"express":    "https://github.com/expressjs/express",
		"react":      "https://github.com/facebook/react",
		"vue":        "https://github.com/vuejs/vue",
		"angular":    "https://github.com/angular/angular",
		"lodash":     "https://github.com/lodash/lodash",
		"axios":      "https://github.com/axios/axios",
		"typescript": "https://github.com/microsoft/TypeScript",
		"webpack":    "https://github.com/webpack/webpack",
		"babel":      "https://github.com/babel/babel",
		"eslint":     "https://github.com/eslint/eslint",
		"prettier":   "https://github.com/prettier/prettier",
		"jest":       "https://github.com/facebook/jest",
		"next":       "https://github.com/vercel/next.js",
		"vite":       "https://github.com/vitejs/vite",
	}

	if url, exists := commonPackages[dep.Name]; exists {
		return url
	}

	// For scoped packages, try the owner/repo pattern
	if strings.HasPrefix(dep.Name, "@") && dep.Owner != "" && dep.Repo != "" {
		return "https://github.com/" + dep.Owner + "/" + dep.Repo
	}

	// Try common patterns for popular packages
	if dep.Repo != "" {
		return "https://github.com/" + dep.Repo + "/" + dep.Repo
	}

	return ""
}
