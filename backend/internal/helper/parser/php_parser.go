package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PHPParser handles parsing of PHP composer.json
type PHPParser struct{}

// NewPHPParser creates a new instance of PHPParser
func NewPHPParser() *PHPParser {
	return &PHPParser{}
}

// GetRuntime returns the runtime type for PHP
func (p *PHPParser) GetRuntime() RuntimeType {
	return RuntimePHP
}

// Parse parses PHP composer.json
func (p *PHPParser) Parse(content string) ([]DependencyInfo, error) {
	var composerJSON struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}

	if err := json.Unmarshal([]byte(content), &composerJSON); err != nil {
		return nil, fmt.Errorf("failed to parse composer.json: %w", err)
	}

	var dependencies []DependencyInfo

	// Parse regular dependencies
	for name, version := range composerJSON.Require {
		if !strings.HasPrefix(name, "php") { // Skip PHP version requirement
			depInfo := p.ParseDependency(name, version)
			dependencies = append(dependencies, *depInfo)
		}
	}

	// Parse dev dependencies
	for name, version := range composerJSON.RequireDev {
		depInfo := p.ParseDependency(name, version)
		dependencies = append(dependencies, *depInfo)
	}

	return dependencies, nil
}

// ParseDependency parses a single PHP dependency
func (p *PHPParser) ParseDependency(name, version string) *DependencyInfo {
	parts := strings.Split(name, "/")
	owner := ""
	repo := name
	if len(parts) >= 2 {
		owner = parts[0]
		repo = parts[1]
	}

	return &DependencyInfo{
		Name:    name,
		Owner:   owner,
		Repo:    repo,
		Version: version,
		Runtime: string(RuntimePHP),
	}
}

// GetRepositoryURL gets GitHub URL for PHP packages
func (p *PHPParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Common PHP packages with known GitHub repositories
	commonPackages := map[string]string{
		"laravel/framework":       "https://github.com/laravel/framework",
		"symfony/symfony":         "https://github.com/symfony/symfony",
		"doctrine/orm":            "https://github.com/doctrine/orm",
		"monolog/monolog":         "https://github.com/Seldaek/monolog",
		"guzzlehttp/guzzle":       "https://github.com/guzzle/guzzle",
		"phpunit/phpunit":         "https://github.com/sebastianbergmann/phpunit",
		"twig/twig":               "https://github.com/twigphp/Twig",
		"swiftmailer/swiftmailer": "https://github.com/swiftmailer/swiftmailer",
		"intervention/image":      "https://github.com/Intervention/image",
		"league/flysystem":        "https://github.com/thephpleague/flysystem",
	}

	if url, exists := commonPackages[dep.Name]; exists {
		return url
	}

	// PHP packages often follow vendor/package pattern
	if strings.Contains(dep.Name, "/") && dep.Owner != "" && dep.Repo != "" {
		return "https://github.com/" + dep.Owner + "/" + dep.Repo
	}

	return ""
}
