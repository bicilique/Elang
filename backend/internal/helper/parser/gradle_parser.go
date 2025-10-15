package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// GradleParser handles parsing of Gradle build files
type GradleParser struct{}

// NewGradleParser creates a new instance of GradleParser
func NewGradleParser() *GradleParser {
	return &GradleParser{}
}

// GetRuntime returns the runtime type for Gradle
func (p *GradleParser) GetRuntime() RuntimeType {
	return RuntimeGradle
}

// Parse parses build.gradle and build.gradle.kts files
func (p *GradleParser) Parse(content string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo

	// Regex patterns for Gradle dependency parsing
	// Handle various Gradle dependency configurations: implementation, api, compile, etc.
	// Pattern 1: Single line dependencies like: implementation 'group:artifact:version'
	singleLineRegex := regexp.MustCompile(`(?m)^\s*(implementation|api|compile|testImplementation|androidTestImplementation|debugImplementation|releaseImplementation|compileOnly|runtimeOnly|annotationProcessor|kapt|ksp)\s+['"]([\w\.-]+):([\w\.-]+):([^'"]+)['"]`)

	// Pattern 2: Dependencies with configurations like: implementation('group:artifact:version') { ... }
	configBlockRegex := regexp.MustCompile(`(?m)^\s*(implementation|api|compile|testImplementation|androidTestImplementation|debugImplementation|releaseImplementation|compileOnly|runtimeOnly|annotationProcessor|kapt|ksp)\s*\(\s*['"]([\w\.-]+):([\w\.-]+):([^'"]+)['"]\s*\)`)

	// Pattern 3: Platform/BOM dependencies like: implementation platform('group:artifact:version')
	platformRegex := regexp.MustCompile(`(?m)^\s*(implementation|api|compile|testImplementation|androidTestImplementation|debugImplementation|releaseImplementation|compileOnly|runtimeOnly)\s+platform\s*\(\s*['"]([\w\.-]+):([\w\.-]+):([^'"]+)['"]\s*\)`)

	// Pattern 4: Variable-based versions like: implementation "group:artifact:$version"
	variableVersionRegex := regexp.MustCompile(`(?m)^\s*(implementation|api|compile|testImplementation|androidTestImplementation|debugImplementation|releaseImplementation|compileOnly|runtimeOnly|annotationProcessor|kapt|ksp)\s+["']([\w\.-]+):([\w\.-]+):\$(\w+)["']`)

	// Pattern 5: Project dependencies like: implementation project(':module')
	projectRegex := regexp.MustCompile(`(?m)^\s*(implementation|api|compile|testImplementation|androidTestImplementation|debugImplementation|releaseImplementation|compileOnly|runtimeOnly)\s+project\s*\(\s*['"]:([^'"]+)['"]\s*\)`)

	// Parse single line dependencies
	matches := singleLineRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 5 {
			groupId := strings.TrimSpace(match[2])
			artifactId := strings.TrimSpace(match[3])
			version := strings.TrimSpace(match[4])

			depInfo := p.ParseDependency(fmt.Sprintf("%s:%s", groupId, artifactId), version)
			depInfo.Owner = groupId
			depInfo.Repo = artifactId
			dependencies = append(dependencies, *depInfo)
		}
	}

	// Parse dependencies with configuration blocks
	configMatches := configBlockRegex.FindAllStringSubmatch(content, -1)
	for _, match := range configMatches {
		if len(match) >= 5 {
			groupId := strings.TrimSpace(match[2])
			artifactId := strings.TrimSpace(match[3])
			version := strings.TrimSpace(match[4])

			depInfo := p.ParseDependency(fmt.Sprintf("%s:%s", groupId, artifactId), version)
			depInfo.Owner = groupId
			depInfo.Repo = artifactId
			dependencies = append(dependencies, *depInfo)
		}
	}

	// Parse platform dependencies
	platformMatches := platformRegex.FindAllStringSubmatch(content, -1)
	for _, match := range platformMatches {
		if len(match) >= 5 {
			groupId := strings.TrimSpace(match[2])
			artifactId := strings.TrimSpace(match[3])
			version := strings.TrimSpace(match[4])

			depInfo := p.ParseDependency(fmt.Sprintf("%s:%s", groupId, artifactId), version)
			depInfo.Owner = groupId
			depInfo.Repo = artifactId
			dependencies = append(dependencies, *depInfo)
		}
	}

	// Parse variable-based versions (we'll store the variable name as version)
	variableMatches := variableVersionRegex.FindAllStringSubmatch(content, -1)
	for _, match := range variableMatches {
		if len(match) >= 5 {
			groupId := strings.TrimSpace(match[2])
			artifactId := strings.TrimSpace(match[3])
			variable := strings.TrimSpace(match[4])

			depInfo := p.ParseDependency(fmt.Sprintf("%s:%s", groupId, artifactId), fmt.Sprintf("$%s", variable))
			depInfo.Owner = groupId
			depInfo.Repo = artifactId
			dependencies = append(dependencies, *depInfo)
		}
	}

	// Parse project dependencies
	projectMatches := projectRegex.FindAllStringSubmatch(content, -1)
	for _, match := range projectMatches {
		if len(match) >= 3 {
			moduleName := strings.TrimSpace(match[2])

			depInfo := p.ParseDependency(fmt.Sprintf("project:%s", moduleName), "local")
			depInfo.Owner = "" // Local project, no owner
			depInfo.Repo = moduleName
			dependencies = append(dependencies, *depInfo)
		}
	}

	return dependencies, nil
}

// ParseDependency parses a single Gradle dependency
func (p *GradleParser) ParseDependency(name, version string) *DependencyInfo {
	// Extract groupId and artifactId if in format groupId:artifactId
	parts := strings.Split(name, ":")
	var owner, repo string

	if len(parts) >= 2 {
		if parts[0] == "project" {
			owner = ""
			repo = parts[1]
		} else {
			owner = parts[0]
			repo = parts[1]
		}
	} else {
		owner = ""
		repo = name
	}

	return &DependencyInfo{
		Name:    name,
		Owner:   owner,
		Repo:    repo,
		Version: version,
		Runtime: string(RuntimeGradle),
	}
}

// GetRepositoryURL gets GitHub URL for Gradle packages (same as Java)
func (p *GradleParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Use same logic as Java parser since Gradle typically uses Maven repositories
	javaParser := NewJavaParser()
	return javaParser.GetRepositoryURL(dep)
}
