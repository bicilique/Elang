package parser

import (
	"regexp"
)

// RubyParser handles parsing of Ruby Gemfile
type RubyParser struct{}

// NewRubyParser creates a new instance of RubyParser
func NewRubyParser() *RubyParser {
	return &RubyParser{}
}

// GetRuntime returns the runtime type for Ruby
func (p *RubyParser) GetRuntime() RuntimeType {
	return RuntimeRuby
}

// Parse parses Ruby Gemfile
func (p *RubyParser) Parse(content string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo

	gemRegex := regexp.MustCompile(`gem\s+['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]+)['"])?`)
	matches := gemRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		version := ""
		if len(match) >= 3 && match[2] != "" {
			version = match[2]
		}

		depInfo := p.ParseDependency(match[1], version)
		dependencies = append(dependencies, *depInfo)
	}

	return dependencies, nil
}

// ParseDependency parses a single Ruby dependency
func (p *RubyParser) ParseDependency(name, version string) *DependencyInfo {
	return &DependencyInfo{
		Name:    name,
		Owner:   "", // Would need RubyGems API to determine
		Repo:    name,
		Version: version,
		Runtime: string(RuntimeRuby),
	}
}

// GetRepositoryURL gets GitHub URL for Ruby gems
func (p *RubyParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Common Ruby gems with known GitHub repositories
	commonGems := map[string]string{
		"rails":        "https://github.com/rails/rails",
		"devise":       "https://github.com/heartcombo/devise",
		"puma":         "https://github.com/puma/puma",
		"sidekiq":      "https://github.com/mperham/sidekiq",
		"rspec":        "https://github.com/rspec/rspec",
		"capybara":     "https://github.com/teamcapybara/capybara",
		"factory_bot":  "https://github.com/thoughtbot/factory_bot",
		"faker":        "https://github.com/faker-ruby/faker",
		"rubocop":      "https://github.com/rubocop/rubocop",
		"activerecord": "https://github.com/rails/rails",
		"actionpack":   "https://github.com/rails/rails",
		"nokogiri":     "https://github.com/sparklemotion/nokogiri",
		"carrierwave":  "https://github.com/carrierwaveuploader/carrierwave",
		"paperclip":    "https://github.com/thoughtbot/paperclip",
	}

	if url, exists := commonGems[dep.Name]; exists {
		return url
	}

	// Try common patterns
	if dep.Repo != "" {
		return "https://github.com/" + dep.Repo + "/" + dep.Repo
	}

	return ""
}
