package helper

import (
	"elang-backend/internal/helper/parser"
	"regexp"

	"strings"
)

// DependencyNameNormalizer handles normalization of dependency names for CVE checking
type DependencyNameNormalizer struct{}

// NewDependencyNameNormalizer creates a new dependency name normalizer
func NewDependencyNameNormalizer() *DependencyNameNormalizer {
	return &DependencyNameNormalizer{}
}

// NormalizeDependencyInfo normalizes a DependencyInfo for CVE checking
func (n *DependencyNameNormalizer) NormalizeDependencyInfo(dep parser.DependencyInfo) parser.DependencyInfo {
	normalized := dep // Copy the struct
	normalized.Name = n.NormalizeName(dep.Name, dep.Runtime)
	normalized.Version = n.NormalizeVersion(dep.Version)
	return normalized
}

// NormalizeName normalizes dependency names based on runtime for CVE database compatibility
func (n *DependencyNameNormalizer) NormalizeName(name, runtime string) string {
	switch strings.ToLower(runtime) {
	case "go":
		return n.normalizeGoName(name)
	case "node", "npm":
		return n.normalizeNodeName(name)
	case "python", "pip":
		return n.normalizePythonName(name)
	case "java", "maven":
		return n.normalizeJavaName(name)
	case "gradle":
		return n.normalizeGradleName(name)
	case "dotnet", "nuget":
		return n.normalizeDotNetName(name)
	case "ruby", "gem":
		return n.normalizeRubyName(name)
	case "php", "composer":
		return n.normalizePHPName(name)
	case "rust", "cargo":
		return n.normalizeRustName(name)
	default:
		return strings.TrimSpace(name)
	}
}

// normalizeGoName normalizes Go module names for OSV compatibility
func (n *DependencyNameNormalizer) normalizeGoName(name string) string {
	// Go modules in OSV use full import paths
	name = strings.TrimSpace(name)

	// Remove version suffixes like /v2, /v3, etc.
	versionSuffixRegex := regexp.MustCompile(`/v\d+$`)
	name = versionSuffixRegex.ReplaceAllString(name, "")

	// Ensure github.com packages are properly formatted
	if strings.Contains(name, "github.com") && !strings.HasPrefix(name, "github.com/") {
		parts := strings.Split(name, "/")
		for i, part := range parts {
			if part == "github.com" && i+2 < len(parts) {
				return strings.Join(parts[i:i+3], "/")
			}
		}
	}

	return name
}

// normalizeNodeName normalizes npm package names for OSV compatibility
func (n *DependencyNameNormalizer) normalizeNodeName(name string) string {
	name = strings.TrimSpace(name)

	// Remove version information if present in the name
	// Examples: @babel/core@7.22.9 -> @babel/core
	if strings.Contains(name, "@") && !strings.HasPrefix(name, "@") {
		// Not a scoped package, so @ indicates version
		name = strings.Split(name, "@")[0]
	} else if strings.HasPrefix(name, "@") {
		// Scoped package, check for version after the package name
		parts := strings.Split(name, "@")
		if len(parts) > 2 { // @scope/package@version
			name = parts[0] + "@" + parts[1] // @scope/package
		}
	}

	// OSV expects exact npm package names, including scoped packages
	// Examples:
	// - express -> express
	// - @babel/core -> @babel/core
	// - @types/node -> @types/node

	return strings.TrimSpace(name)
}

// normalizePythonName normalizes Python package names for PyPI compatibility
func (n *DependencyNameNormalizer) normalizePythonName(name string) string {
	name = strings.TrimSpace(name)

	// PyPI package names are case-insensitive and can contain hyphens or underscores
	// OSV typically uses the canonical form (lowercase with hyphens)
	name = strings.ToLower(name)

	// Convert underscores to hyphens (PyPI canonical form)
	name = strings.ReplaceAll(name, "_", "-")

	// Handle common variations
	commonMappings := map[string]string{
		"pil":             "pillow",
		"yaml":            "pyyaml",
		"psycopg2":        "psycopg2-binary", // Handle both forms
		"psycopg2-binary": "psycopg2-binary",
	}

	if canonical, exists := commonMappings[name]; exists {
		return canonical
	}

	return name
}

// normalizeJavaName normalizes Java/Maven artifact names for OSV compatibility
func (n *DependencyNameNormalizer) normalizeJavaName(name string) string {
	name = strings.TrimSpace(name)

	// Remove version information if present
	// Examples: org.springframework:spring-core:5.3.21 -> org.springframework:spring-core
	if strings.Count(name, ":") >= 2 {
		parts := strings.Split(name, ":")
		if len(parts) >= 3 {
			// groupId:artifactId:version -> groupId:artifactId
			name = parts[0] + ":" + parts[1]
		}
	}

	// Maven artifacts in OSV use groupId:artifactId format
	// Examples:
	// - org.springframework:spring-core
	// - com.fasterxml.jackson.core:jackson-core

	// If it already has groupId:artifactId format, return as-is
	if strings.Contains(name, ":") {
		return name
	}

	// Try to infer common group IDs for well-known artifacts
	commonGroupIds := map[string]string{
		"spring-core":          "org.springframework:spring-core",
		"spring-boot":          "org.springframework.boot:spring-boot",
		"hibernate-core":       "org.hibernate:hibernate-core",
		"jackson-core":         "com.fasterxml.jackson.core:jackson-core",
		"jackson-databind":     "com.fasterxml.jackson.core:jackson-databind",
		"commons-lang3":        "org.apache.commons:commons-lang3",
		"commons-collections4": "org.apache.commons:commons-collections4",
		"guava":                "com.google.guava:guava",
		"slf4j-api":            "org.slf4j:slf4j-api",
		"logback-classic":      "ch.qos.logback:logback-classic",
		"junit-jupiter":        "org.junit.jupiter:junit-jupiter",
	}

	if fullName, exists := commonGroupIds[name]; exists {
		return fullName
	}

	return name
}

// normalizeGradleName normalizes Gradle dependency names (uses Maven ecosystem)
func (n *DependencyNameNormalizer) normalizeGradleName(name string) string {
	// Gradle uses the same format as Maven for OSV
	return n.normalizeJavaName(name)
}

// normalizeDotNetName normalizes .NET/NuGet package names for OSV compatibility
func (n *DependencyNameNormalizer) normalizeDotNetName(name string) string {
	name = strings.TrimSpace(name)

	// Remove version paths if present
	// Examples: Microsoft.AspNetCore.App/6.0.0 -> Microsoft.AspNetCore.App
	if strings.Contains(name, "/") {
		name = strings.Split(name, "/")[0]
	}

	// NuGet package names are case-insensitive but OSV typically uses PascalCase
	// Examples:
	// - Newtonsoft.Json
	// - Microsoft.AspNetCore.Mvc
	// - System.Text.Json

	// Convert to proper casing for well-known packages
	commonMappings := map[string]string{
		"newtonsoft.json":          "Newtonsoft.Json",
		"microsoft.aspnetcore":     "Microsoft.AspNetCore",
		"microsoft.aspnetcore.mvc": "Microsoft.AspNetCore.Mvc",
		"microsoft.aspnetcore.app": "Microsoft.AspNetCore.App",
		"system.text.json":         "System.Text.Json",
		"entityframework":          "EntityFramework",
		"automapper":               "AutoMapper",
		"serilog":                  "Serilog",
		"fluentvalidation":         "FluentValidation",
	}

	lowerName := strings.ToLower(name)
	if canonical, exists := commonMappings[lowerName]; exists {
		return canonical
	}

	// Default: return original name (NuGet is case-insensitive)
	return name
}

// normalizeRubyName normalizes Ruby gem names for RubyGems compatibility
func (n *DependencyNameNormalizer) normalizeRubyName(name string) string {
	name = strings.TrimSpace(name)

	// Ruby gem names are typically lowercase with hyphens or underscores
	// OSV typically uses the canonical gem name
	name = strings.ToLower(name)

	// Common mappings for Ruby gems
	commonMappings := map[string]string{
		"rails":          "rails",
		"devise":         "devise",
		"sidekiq":        "sidekiq",
		"rspec":          "rspec",
		"puma":           "puma",
		"bootsnap":       "bootsnap",
		"turbo-rails":    "turbo-rails",
		"stimulus-rails": "stimulus-rails",
	}

	if canonical, exists := commonMappings[name]; exists {
		return canonical
	}

	return name
}

// normalizePHPName normalizes PHP/Composer package names for Packagist compatibility
func (n *DependencyNameNormalizer) normalizePHPName(name string) string {
	name = strings.TrimSpace(name)

	// Composer packages use vendor/package format
	// Examples:
	// - symfony/symfony
	// - laravel/framework
	// - guzzlehttp/guzzle

	name = strings.ToLower(name)

	// If it already has vendor/package format, return as-is
	if strings.Contains(name, "/") {
		return name
	}

	// Try to infer vendor for well-known packages
	commonVendors := map[string]string{
		"symfony":     "symfony/symfony",
		"laravel":     "laravel/framework",
		"guzzle":      "guzzlehttp/guzzle",
		"twig":        "twig/twig",
		"monolog":     "monolog/monolog",
		"phpunit":     "phpunit/phpunit",
		"doctrine":    "doctrine/orm",
		"swiftmailer": "swiftmailer/swiftmailer",
	}

	if fullName, exists := commonVendors[name]; exists {
		return fullName
	}

	return name
}

// normalizeRustName normalizes Rust crate names for crates.io compatibility
func (n *DependencyNameNormalizer) normalizeRustName(name string) string {
	name = strings.TrimSpace(name)

	// Rust crate names are typically lowercase with hyphens or underscores
	// OSV uses the exact crate name from crates.io
	name = strings.ToLower(name)

	// Convert underscores to hyphens (canonical form on crates.io)
	name = strings.ReplaceAll(name, "_", "-")

	return name
}

// NormalizeVersion normalizes version strings for consistency
func (n *DependencyNameNormalizer) NormalizeVersion(version string) string {
	version = strings.TrimSpace(version)

	// Remove common version prefixes and constraints
	version = strings.TrimPrefix(version, "^")
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, ">=")
	version = strings.TrimPrefix(version, "<=")
	version = strings.TrimPrefix(version, ">")
	version = strings.TrimPrefix(version, "<")
	version = strings.TrimPrefix(version, "=")
	version = strings.TrimPrefix(version, "v")

	// Handle version ranges - take the first version
	if strings.Contains(version, " - ") {
		parts := strings.Split(version, " - ")
		version = strings.TrimSpace(parts[0])
	}

	// Handle OR conditions - take the first version
	if strings.Contains(version, " || ") {
		parts := strings.Split(version, " || ")
		version = strings.TrimSpace(parts[0])
	}

	// Remove any trailing spaces or comments
	if idx := strings.Index(version, " "); idx != -1 {
		version = version[:idx]
	}

	return strings.TrimSpace(version)
}

// NormalizeBatch normalizes a batch of dependencies
func (n *DependencyNameNormalizer) NormalizeBatch(dependencies []parser.DependencyInfo) []parser.DependencyInfo {
	normalized := make([]parser.DependencyInfo, 0, len(dependencies))

	for _, dep := range dependencies {
		normalized = append(normalized, n.NormalizeDependencyInfo(dep))
	}

	return normalized
}

// ValidateForCVECheck validates if a dependency is suitable for CVE checking
func (n *DependencyNameNormalizer) ValidateForCVECheck(dep parser.DependencyInfo) bool {
	// Check if name is not empty
	if strings.TrimSpace(dep.Name) == "" {
		return false
	}

	// Check if version is not empty
	if strings.TrimSpace(dep.Version) == "" {
		return false
	}

	// Check if runtime is supported
	supportedRuntimes := map[string]bool{
		"go":       true,
		"node":     true,
		"npm":      true,
		"python":   true,
		"pip":      true,
		"java":     true,
		"maven":    true,
		"gradle":   true,
		"dotnet":   true,
		"nuget":    true,
		"ruby":     true,
		"gem":      true,
		"php":      true,
		"composer": true,
		"rust":     true,
		"cargo":    true,
	}

	return supportedRuntimes[strings.ToLower(dep.Runtime)]
}

// GetCVECompatibleName returns the CVE-database compatible name for a dependency
func (n *DependencyNameNormalizer) GetCVECompatibleName(dep parser.DependencyInfo) string {
	return n.NormalizeName(dep.Name, dep.Runtime)
}

// GetSuggestedNames returns alternative name suggestions for CVE checking
func (n *DependencyNameNormalizer) GetSuggestedNames(dep parser.DependencyInfo) []string {
	suggestions := []string{}
	baseName := n.NormalizeName(dep.Name, dep.Runtime)
	suggestions = append(suggestions, baseName)

	switch strings.ToLower(dep.Runtime) {
	case "python", "pip":
		// Add both hyphen and underscore variations
		if strings.Contains(baseName, "-") {
			suggestions = append(suggestions, strings.ReplaceAll(baseName, "-", "_"))
		}
		if strings.Contains(baseName, "_") {
			suggestions = append(suggestions, strings.ReplaceAll(baseName, "_", "-"))
		}

	case "rust", "cargo":
		// Add both hyphen and underscore variations
		if strings.Contains(baseName, "-") {
			suggestions = append(suggestions, strings.ReplaceAll(baseName, "-", "_"))
		}
		if strings.Contains(baseName, "_") {
			suggestions = append(suggestions, strings.ReplaceAll(baseName, "_", "-"))
		}

	case "dotnet", "nuget":
		// Add case variations
		lowerName := strings.ToLower(baseName)
		if lowerName != baseName {
			suggestions = append(suggestions, lowerName)
		}

		// Add title case variation (manual implementation since strings.Title is deprecated)
		titleName := toTitleCase(baseName)
		if titleName != baseName && titleName != lowerName {
			suggestions = append(suggestions, titleName)
		}
	}

	// Remove duplicates
	unique := make(map[string]bool)
	var result []string
	for _, suggestion := range suggestions {
		if !unique[suggestion] {
			unique[suggestion] = true
			result = append(result, suggestion)
		}
	}

	return result
}

// toTitleCase converts a string to title case, handling dots properly
func toTitleCase(s string) string {
	if s == "" {
		return s
	}

	// Split by dots and title case each part
	parts := strings.Split(s, ".")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, ".")
}
