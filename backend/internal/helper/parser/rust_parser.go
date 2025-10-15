package parser

import (
	"regexp"
	"strings"
)

// RustParser handles parsing of Rust Cargo.toml
type RustParser struct{}

// NewRustParser creates a new instance of RustParser
func NewRustParser() *RustParser {
	return &RustParser{}
}

// GetRuntime returns the runtime type for Rust
func (p *RustParser) GetRuntime() RuntimeType {
	return RuntimeRust
}

// Parse parses Rust Cargo.toml
func (p *RustParser) Parse(content string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo

	// Find [dependencies] and [dev-dependencies] sections
	// Use a more precise regex that looks for section headers at line start
	sectionPattern := regexp.MustCompile(`(?m)^\[((?:dev-)?dependencies)\]\s*$`)
	matches := sectionPattern.FindAllStringSubmatchIndex(content, -1)

	// Also find any other section to determine boundaries
	anySectionPattern := regexp.MustCompile(`(?m)^\[.*?\]\s*$`)
	allSections := anySectionPattern.FindAllStringIndex(content, -1)

	for _, match := range matches {
		sectionStart := match[1] // End of section name
		var sectionEnd int

		// Find the next section after this one
		nextSectionPos := len(content)
		for _, otherSection := range allSections {
			if otherSection[0] > match[0] && otherSection[0] < nextSectionPos {
				nextSectionPos = otherSection[0]
			}
		}
		sectionEnd = nextSectionPos

		sectionContent := content[sectionStart:sectionEnd]

		// Parse dependencies in this section
		deps := p.parseCargoSection(sectionContent)
		dependencies = append(dependencies, deps...)
	}

	return dependencies, nil
}

// parseCargoSection parses a single dependency section from Cargo.toml
func (p *RustParser) parseCargoSection(content string) []DependencyInfo {
	var dependencies []DependencyInfo

	// Split into lines and process
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse dependency line
		if dep := p.parseCargoLine(line); dep != nil {
			dependencies = append(dependencies, *dep)
		}
	}

	return dependencies
}

// parseCargoLine parses a single dependency line from Cargo.toml
func (p *RustParser) parseCargoLine(line string) *DependencyInfo {
	// Pattern 1: simple string version
	// serde_json = "1.0.104"
	simplePattern := regexp.MustCompile(`^([a-zA-Z0-9\-_]+)\s*=\s*"([^"]+)"`)
	if matches := simplePattern.FindStringSubmatch(line); len(matches) >= 3 {
		return p.ParseDependency(matches[1], matches[2])
	}

	// Pattern 2: object with version (single line)
	// tokio = { version = "1.29.1", features = ["full"] }
	objectPattern := regexp.MustCompile(`^([a-zA-Z0-9\-_]+)\s*=\s*\{.*?version\s*=\s*"([^"]+)"`)
	if matches := objectPattern.FindStringSubmatch(line); len(matches) >= 3 {
		return p.ParseDependency(matches[1], matches[2])
	}

	// Pattern 3: multi-line object start - check if it contains version on same line
	if strings.Contains(line, "=") && strings.Contains(line, "{") {
		namePattern := regexp.MustCompile(`^([a-zA-Z0-9\-_]+)\s*=`)
		nameMatches := namePattern.FindStringSubmatch(line)
		if len(nameMatches) >= 2 {
			name := nameMatches[1]

			// Look for version anywhere in the line
			versionPattern := regexp.MustCompile(`version\s*=\s*"([^"]+)"`)
			if versionMatches := versionPattern.FindStringSubmatch(line); len(versionMatches) >= 2 {
				return p.ParseDependency(name, versionMatches[1])
			}
		}
	}

	return nil
}

// ParseDependency parses a single Rust dependency
func (p *RustParser) ParseDependency(name, version string) *DependencyInfo {
	return &DependencyInfo{
		Name:    name,
		Owner:   "",
		Repo:    name,
		Version: version,
		Runtime: string(RuntimeRust),
	}
}

// GetRepositoryURL gets GitHub URL for Rust crates
func (p *RustParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Common Rust crates with known GitHub repositories
	commonCrates := map[string]string{
		"tokio":       "https://github.com/tokio-rs/tokio",
		"serde":       "https://github.com/serde-rs/serde",
		"reqwest":     "https://github.com/seanmonstar/reqwest",
		"clap":        "https://github.com/clap-rs/clap",
		"anyhow":      "https://github.com/dtolnay/anyhow",
		"thiserror":   "https://github.com/dtolnay/thiserror",
		"syn":         "https://github.com/dtolnay/syn",
		"quote":       "https://github.com/dtolnay/quote",
		"proc-macro2": "https://github.com/dtolnay/proc-macro2",
		"futures":     "https://github.com/rust-lang/futures-rs",
		"regex":       "https://github.com/rust-lang/regex",
		"rand":        "https://github.com/rust-random/rand",
		"chrono":      "https://github.com/chronotope/chrono",
		"uuid":        "https://github.com/uuid-rs/uuid",
		"log":         "https://github.com/rust-lang/log",
		"env_logger":  "https://github.com/rust-cli/env_logger",
	}

	if url, exists := commonCrates[dep.Name]; exists {
		return url
	}

	// Try common patterns
	if dep.Repo != "" {
		return "https://github.com/" + dep.Repo + "/" + dep.Repo
	}

	return ""
}
