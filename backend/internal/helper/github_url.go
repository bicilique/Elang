package helper

import (
	"fmt"
	"regexp"
	"strings"
)

type GitHubRepoParts struct {
	Owner string
	Repo  string
}

// ExtractGitHubOwnerRepo extracts the owner and repo from a GitHub URL.
// Example: https://github.com/gin-gonic/gin -> gin-gonic, gin
func ExtractGitHubOwnerRepo(url string) (GitHubRepoParts, bool) {
	// Remove trailing .git or slashes
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimRight(url, "/")

	// Regex for https://github.com/owner/repo (with or without www)
	re := regexp.MustCompile(`(?i)^https?://(www\.)?github\.com/([^/]+)/([^/]+)$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 4 {
		return GitHubRepoParts{Owner: matches[2], Repo: matches[3]}, true
	}
	return GitHubRepoParts{}, false
}

// GetCommitSHAFromVersion finds the commit SHA for a given version from a list of tags.
// It supports exact, v-prefix, .x, and pre-release suffixes (dev/rc/alpha/beta, etc.).
func GetCommitSHAFromVersion(version string, tags []map[string]interface{}) (string, bool) {
	versionLower := strings.ToLower(strings.TrimSpace(version))
	if versionLower == "" {
		return "", false
	}
	// Try exact match first (with and without v prefix)
	for _, tag := range tags {
		tagName, _ := tag["name"].(string)
		tagNameLower := strings.ToLower(strings.TrimSpace(tagName))
		if tagNameLower == versionLower ||
			tagNameLower == "v"+versionLower ||
			strings.TrimPrefix(tagNameLower, "v") == versionLower {
			sha, shaOk := tag["commit_sha"].(string)
			if shaOk && sha != "" {
				return sha, true
			}
		}
	}
	// Fuzzy match: handle .x, dev, rc, alpha, beta, etc.
	for _, tag := range tags {
		tagName, _ := tag["name"].(string)
		tagNameLower := strings.ToLower(strings.TrimSpace(tagName))
		// Pre-release: match v9.15.0-beta, v9.15.0-beta.1, etc.
		if strings.HasPrefix(tagNameLower, versionLower+"-") ||
			strings.HasPrefix(tagNameLower, "v"+versionLower+"-") {
			sha, shaOk := tag["commit_sha"].(string)
			if shaOk && sha != "" {
				return sha, true
			}
		}
		// Pre-release: match v9.15.0-beta, v9.15.0-beta.1, etc. if versionLower already has -beta, -rc, etc.
		if strings.HasPrefix(tagNameLower, versionLower) ||
			strings.HasPrefix(tagNameLower, "v"+versionLower) {
			sha, shaOk := tag["commit_sha"].(string)
			if shaOk && sha != "" {
				return sha, true
			}
		}
		// Handle .x (e.g., 1.2.x matches 1.2.3, 1.2.0, etc.)
		if strings.HasSuffix(versionLower, ".x") {
			prefix := strings.TrimSuffix(versionLower, ".x")
			if strings.HasPrefix(tagNameLower, prefix) || strings.HasPrefix(strings.TrimPrefix(tagNameLower, "v"), prefix) {
				sha, shaOk := tag["commit_sha"].(string)
				if shaOk && sha != "" {
					return sha, true
				}
			}
		}
		// Fuzzy: allow versionLower to match tagNameLower with -beta, -rc, -alpha, etc. suffixes
		if strings.HasPrefix(tagNameLower, versionLower+"-") ||
			strings.HasPrefix(tagNameLower, versionLower+".") ||
			strings.HasPrefix(tagNameLower, "v"+versionLower+"-") ||
			strings.HasPrefix(tagNameLower, "v"+versionLower+".") {
			sha, shaOk := tag["commit_sha"].(string)
			if shaOk && sha != "" {
				return sha, true
			}
		}
	}
	return "", false // Not found
}

// NormalizeVersion removes common prefixes from version strings to allow better matching.
// It handles prefixes like "v", "r", "release-", "version-", etc.
func NormalizeVersion(version string) string {
	if version == "" {
		return ""
	}

	// Trim whitespace first
	normalized := strings.TrimSpace(version)

	// Remove common prefixes (case-insensitive)
	prefixes := []string{
		"version-",
		"version_",
		"release-",
		"release_",
		"tag-",
		"tag_",
		"v",
		"r",
	}

	// Convert to lowercase for comparison, but preserve original case for result
	lowerNormalized := strings.ToLower(normalized)

	for _, prefix := range prefixes {
		lowerPrefix := strings.ToLower(prefix)
		if strings.HasPrefix(lowerNormalized, lowerPrefix) {
			// Remove the prefix from the original (preserving case)
			normalized = normalized[len(prefix):]
			break // Only remove the first matching prefix
		}
	}

	return strings.ToLower(strings.TrimSpace(normalized))
}

// VersionsMatch checks if two version strings represent the same version,
// handling common prefixes and case differences.
func VersionsMatch(version1, version2 string) bool {
	if version1 == version2 {
		return true // Exact match
	}

	norm1 := NormalizeVersion(version1)
	norm2 := NormalizeVersion(version2)

	return norm1 == norm2 && norm1 != ""
}

// FindBestMatchingTag finds the best matching tag for a given version from a list of tags.
// It returns the actual tag name as it exists in the repository, or empty string if not found.
func FindBestMatchingTag(version string, tags []map[string]interface{}) string {
	if version == "" || len(tags) == 0 {
		return ""
	}

	// First try exact match
	for _, tag := range tags {
		if tagName, ok := tag["name"].(string); ok {
			if tagName == version {
				return tagName
			}
		}
	}

	// Then try normalized matching
	normalizedVersion := NormalizeVersion(version)
	for _, tag := range tags {
		if tagName, ok := tag["name"].(string); ok {
			if NormalizeVersion(tagName) == normalizedVersion {
				return tagName
			}
		}
	}

	return ""
}

// ValidateTagsExist checks if both base and head tags exist in the repository.
// Returns the actual tag names as they exist in the repository, or error if not found.
func ValidateTagsExist(baseVersion, headVersion string, tags []map[string]interface{}) (string, string, error) {
	baseTag := FindBestMatchingTag(baseVersion, tags)
	if baseTag == "" {
		return "", "", fmt.Errorf("base tag not found for version: %s", baseVersion)
	}

	headTag := FindBestMatchingTag(headVersion, tags)
	if headTag == "" {
		return "", "", fmt.Errorf("head tag not found for version: %s", headVersion)
	}

	return baseTag, headTag, nil
}
