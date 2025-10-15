package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// JavaParser handles parsing of Java Maven pom.xml files
type JavaParser struct{}

// NewJavaParser creates a new instance of JavaParser
func NewJavaParser() *JavaParser {
	return &JavaParser{}
}

// GetRuntime returns the runtime type for Java
func (p *JavaParser) GetRuntime() RuntimeType {
	return RuntimeJava
}

// Parse parses Maven pom.xml files
func (p *JavaParser) Parse(content string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo

	// Enhanced regex patterns for XML parsing
	// Handle multi-line dependency blocks with more flexible whitespace
	dependencyBlock := regexp.MustCompile(`(?s)<dependency[^>]*>\s*<groupId>\s*([^<\s]+)\s*</groupId>\s*<artifactId>\s*([^<\s]+)\s*</artifactId>\s*(?:<version>\s*([^<\s]+)\s*</version>\s*)?(?:<scope>[^<]*</scope>\s*)?</dependency>`)

	matches := dependencyBlock.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			groupId := strings.TrimSpace(match[1])
			artifactId := strings.TrimSpace(match[2])
			version := ""
			if len(match) >= 4 && match[3] != "" {
				version = strings.TrimSpace(match[3])
			}

			depInfo := p.ParseDependency(fmt.Sprintf("%s:%s", groupId, artifactId), version)
			depInfo.Owner = groupId
			depInfo.Repo = artifactId
			dependencies = append(dependencies, *depInfo)
		}
	}

	return dependencies, nil
}

// ParseDependency parses a single Java dependency
func (p *JavaParser) ParseDependency(name, version string) *DependencyInfo {
	// Extract groupId and artifactId if in format groupId:artifactId
	parts := strings.Split(name, ":")
	var owner, repo string

	if len(parts) >= 2 {
		owner = parts[0]
		repo = parts[1]
	} else {
		owner = ""
		repo = name
	}

	return &DependencyInfo{
		Name:    name,
		Owner:   owner,
		Repo:    repo,
		Version: version,
		Runtime: string(RuntimeJava),
	}
}

// GetRepositoryURL gets GitHub URL for Java packages
func (p *JavaParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Common Java packages with known GitHub repositories
	commonPackages := map[string]string{
		"org.springframework:spring-core":             "https://github.com/spring-projects/spring-framework",
		"org.springframework:spring-boot":             "https://github.com/spring-projects/spring-boot",
		"com.fasterxml.jackson.core:jackson-core":     "https://github.com/FasterXML/jackson-core",
		"com.fasterxml.jackson.core:jackson-databind": "https://github.com/FasterXML/jackson-databind",
		"org.apache.commons:commons-lang3":            "https://github.com/apache/commons-lang",
		"com.google.guava:guava":                      "https://github.com/google/guava",
		"junit:junit":                                 "https://github.com/junit-team/junit4",
		"org.junit.jupiter:junit-jupiter":             "https://github.com/junit-team/junit5",
		"org.mockito:mockito-core":                    "https://github.com/mockito/mockito",
		"org.slf4j:slf4j-api":                         "https://github.com/qos-ch/slf4j",
		"ch.qos.logback:logback-classic":              "https://github.com/qos-ch/logback",
	}

	if url, exists := commonPackages[dep.Name]; exists {
		return url
	}

	// Try to construct from groupId/artifactId
	if dep.Owner != "" && dep.Repo != "" {
		// Handle common organization patterns
		if strings.Contains(dep.Owner, "org.springframework") {
			return "https://github.com/spring-projects/" + dep.Repo
		}
		if strings.Contains(dep.Owner, "com.fasterxml.jackson") {
			return "https://github.com/FasterXML/" + dep.Repo
		}
		if strings.Contains(dep.Owner, "org.apache") {
			return "https://github.com/apache/" + dep.Repo
		}

		// Default pattern
		return "https://github.com/" + dep.Owner + "/" + dep.Repo
	}

	return ""
}
