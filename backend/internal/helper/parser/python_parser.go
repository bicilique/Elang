package parser

import (
	"bufio"
	"regexp"
	"strings"
)

// PythonParser handles parsing of Python dependency files
type PythonParser struct{}

// NewPythonParser creates a new instance of PythonParser
func NewPythonParser() *PythonParser {
	return &PythonParser{}
}

// GetRuntime returns the runtime type for Python
func (p *PythonParser) GetRuntime() RuntimeType {
	return RuntimePython
}

// Parse parses requirements.txt files
func (p *PythonParser) Parse(content string) ([]DependencyInfo, error) {
	var dependencies []DependencyInfo

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip pip directives and URL-based dependencies
		if strings.HasPrefix(line, "-") || strings.Contains(line, "://") {
			continue
		}

		// Remove environment markers (e.g., ; python_version >= "3.8")
		if idx := strings.Index(line, ";"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Handle various requirement formats with complex version specifications
		// Enhanced regex to capture full version specs like >=4.2.0,<5.0
		versionRegex := regexp.MustCompile(`^([a-zA-Z0-9\-_.]+)((?:[><=~!]+[^,;\s]+(?:\s*,\s*[><=~!]+[^,;\s]+)*)).*$`)
		matches := versionRegex.FindStringSubmatch(line)

		if len(matches) >= 3 {
			packageName := matches[1]
			versionSpec := matches[2]

			// Clean version spec by removing operators and keeping only the version number
			cleanVersion := regexp.MustCompile(`^[><=~!]+\s*`).ReplaceAllString(versionSpec, "")
			// For complex specs like ">=4.2.0,<5.0", take the first version
			if idx := strings.Index(cleanVersion, ","); idx != -1 {
				cleanVersion = strings.TrimSpace(cleanVersion[:idx])
			}

			depInfo := p.ParseDependency(packageName, cleanVersion)
			if depInfo != nil {
				dependencies = append(dependencies, *depInfo)
			}
		} else {
			// Just package name without version
			if line != "" {
				depInfo := p.ParseDependency(line, "")
				if depInfo != nil {
					dependencies = append(dependencies, *depInfo)
				}
			}
		}
	}

	return dependencies, nil
}

// ParseDependency parses a single Python dependency
func (p *PythonParser) ParseDependency(name, version string) *DependencyInfo {
	return &DependencyInfo{
		Name:    name,
		Owner:   "", // Would need PyPI API to determine
		Repo:    name,
		Version: version,
		Runtime: string(RuntimePython),
	}
}

// GetRepositoryURL gets GitHub URL for Python packages
func (p *PythonParser) GetRepositoryURL(dep *DependencyInfo) string {
	// Common Python packages with known GitHub repositories
	commonPackages := map[string]string{
		"django":          "https://github.com/django/django",
		"flask":           "https://github.com/pallets/flask",
		"requests":        "https://github.com/psf/requests",
		"numpy":           "https://github.com/numpy/numpy",
		"pandas":          "https://github.com/pandas-dev/pandas",
		"tensorflow":      "https://github.com/tensorflow/tensorflow",
		"pytorch":         "https://github.com/pytorch/pytorch",
		"scikit-learn":    "https://github.com/scikit-learn/scikit-learn",
		"fastapi":         "https://github.com/tiangolo/fastapi",
		"celery":          "https://github.com/celery/celery",
		"pytest":          "https://github.com/pytest-dev/pytest",
		"black":           "https://github.com/psf/black",
		"sqlalchemy":      "https://github.com/sqlalchemy/sqlalchemy",
		"pydantic":        "https://github.com/pydantic/pydantic",
		"pillow":          "https://github.com/python-pillow/Pillow",
		"psycopg2":        "https://github.com/psycopg/psycopg2",
		"psycopg2-binary": "https://github.com/psycopg/psycopg2",
	}

	if url, exists := commonPackages[dep.Name]; exists {
		return url
	}

	// Try common patterns
	if dep.Repo != "" {
		return "https://github.com/" + dep.Repo + "/" + dep.Repo
	}

	return ""
}
