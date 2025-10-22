package helper

import (
	"context"
	"elang-backend/internal/helper/parser"
	"elang-backend/internal/model"

	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"strings"
	"time"
)

// CVEHelper provides vulnerability checking functionality for dependencies
type CVEHelper struct {
	httpClient *http.Client
	timeout    time.Duration
	normalizer *DependencyNameNormalizer
}

// OSVQuery represents the OSV API query structure
type OSVQuery struct {
	Package OSVPackage `json:"package"`
	Version string     `json:"version"`
}

type OSVPackage struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

// OSVResponse represents the OSV API response
type OSVResponse struct {
	Vulns []OSVVulnerability `json:"vulns"`
}

type OSVVulnerability struct {
	ID         string         `json:"id"`
	Summary    string         `json:"summary"`
	Details    string         `json:"details"`
	Affects    []OSVAffected  `json:"affected"`
	References []OSVReference `json:"references"`
}

type OSVAffected struct {
	Package OSVPackage `json:"package"`
	Ranges  []OSVRange `json:"ranges"`
}

type OSVRange struct {
	Type   string     `json:"type"`
	Events []OSVEvent `json:"events"`
}

type OSVEvent struct {
	Introduced string `json:"introduced,omitempty"`
	Fixed      string `json:"fixed,omitempty"`
}

type OSVReference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// NewCVEHelper creates a new CVE helper instance
func NewCVEHelper() *CVEHelper {
	return &CVEHelper{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout:    30 * time.Second,
		normalizer: NewDependencyNameNormalizer(),
	}
}

// CVESeverity represents the severity levels of vulnerabilities
type CVESeverity string

const (
	SeverityCritical CVESeverity = "CRITICAL"
	SeverityHigh     CVESeverity = "HIGH"
	SeverityMedium   CVESeverity = "MEDIUM"
	SeverityLow      CVESeverity = "LOW"
	SeverityInfo     CVESeverity = "INFO"
	SeverityUnknown  CVESeverity = "UNKNOWN"
)

// VulnerabilityInfo represents detailed vulnerability information
type VulnerabilityInfo struct {
	ID                    string      `json:"id"`
	CVE                   string      `json:"cve"`
	Summary               string      `json:"summary"`
	Description           string      `json:"description"`
	Severity              CVESeverity `json:"severity"`
	Score                 float64     `json:"score"`
	AffectedVersions      []string    `json:"affected_versions"`
	PatchedVersions       []string    `json:"patched_versions"`
	References            []string    `json:"references"`
	PublishedDate         time.Time   `json:"published_date"`
	ModifiedDate          time.Time   `json:"modified_date"`
	VectorString          string      `json:"vector_string"`
	AttackVector          string      `json:"attack_vector"`
	AttackComplexity      string      `json:"attack_complexity"`
	PrivilegesRequired    string      `json:"privileges_required"`
	UserInteraction       string      `json:"user_interaction"`
	Scope                 string      `json:"scope"`
	ConfidentialityImpact string      `json:"confidentiality_impact"`
	IntegrityImpact       string      `json:"integrity_impact"`
	AvailabilityImpact    string      `json:"availability_impact"`
	ExploitabilityScore   float64     `json:"exploitability_score"`
	ImpactScore           float64     `json:"impact_score"`
}

// DependencyVulnerabilityResult contains vulnerability results for a dependency
type DependencyVulnerabilityResult struct {
	Dependency      parser.DependencyInfo `json:"dependency"`
	Vulnerabilities []VulnerabilityInfo   `json:"vulnerabilities"`
	IsVulnerable    bool                  `json:"is_vulnerable"`
	TotalCount      int                   `json:"total_count"`
	CriticalCount   int                   `json:"critical_count"`
	HighCount       int                   `json:"high_count"`
	MediumCount     int                   `json:"medium_count"`
	LowCount        int                   `json:"low_count"`
	RiskScore       float64               `json:"risk_score"`
	Recommendations []string              `json:"recommendations"`
	CheckedAt       time.Time             `json:"checked_at"`
	Error           string                `json:"error,omitempty"`
}

// BatchVulnerabilityResult contains results for multiple dependencies
type BatchVulnerabilityResult struct {
	Dependencies           []DependencyVulnerabilityResult `json:"dependencies"`
	TotalDependencies      int                             `json:"total_dependencies"`
	VulnerableDependencies int                             `json:"vulnerable_dependencies"`
	TotalVulnerabilities   int                             `json:"total_vulnerabilities"`
	OverallRiskScore       float64                         `json:"overall_risk_score"`
	HighestSeverity        CVESeverity                     `json:"highest_severity"`
	CheckedAt              time.Time                       `json:"checked_at"`
	Summary                VulnerabilitySummary            `json:"summary"`
}

// VulnerabilitySummary provides aggregated vulnerability statistics
type VulnerabilitySummary struct {
	CriticalCount int `json:"critical_count"`
	HighCount     int `json:"high_count"`
	MediumCount   int `json:"medium_count"`
	LowCount      int `json:"low_count"`
	InfoCount     int `json:"info_count"`
}

// OSVResponse represents the response from OSV API
// Note: This uses the existing OSVResponse and OSVVulnerability types from Security_detector_v2.go

// CheckDependencyVulnerabilities checks vulnerabilities for a single dependency
func (c *CVEHelper) CheckDependencyVulnerabilities(ctx context.Context, dep parser.DependencyInfo) (*DependencyVulnerabilityResult, error) {
	// Normalize the dependency for CVE checking
	normalizedDep := c.normalizer.NormalizeDependencyInfo(dep)

	result := &DependencyVulnerabilityResult{
		Dependency:      normalizedDep, // Use normalized dependency in result
		Vulnerabilities: []VulnerabilityInfo{},
		IsVulnerable:    false,
		CheckedAt:       time.Now(),
	}

	// Only log warnings and errors for important traceability
	if !c.normalizer.ValidateForCVECheck(normalizedDep) {
		result.Error = fmt.Sprintf("Invalid dependency for CVE check: name='%s', version='%s', runtime='%s'",
			normalizedDep.Name, normalizedDep.Version, normalizedDep.Runtime)
		slog.Warn("Invalid dependency for CVE check",
			"name", normalizedDep.Name,
			"version", normalizedDep.Version,
			"runtime", normalizedDep.Runtime)
		return result, nil
	}

	// Check multiple vulnerability databases with alternative names
	osvVulns, err := c.checkOSVDatabase(ctx, normalizedDep)
	if err != nil {
		// Try with alternative names if the primary check failed
		alternatives := c.normalizer.GetSuggestedNames(normalizedDep)
		for _, altName := range alternatives[1:] { // Skip first one as it was already tried
			altDep := normalizedDep
			altDep.Name = altName
			altVulns, altErr := c.checkOSVDatabase(ctx, altDep)
			if altErr == nil && len(altVulns) > 0 {
				// slog.Info("Found vulnerabilities with alternative name",
				// 	"original", normalizedDep.Name,
				// 	"alternative", altName,
				// 	"count", len(altVulns))
				osvVulns = altVulns
				err = nil
				break
			}
		}
	}
	if err != nil {
		slog.Warn("Failed to check OSV database", "dependency", normalizedDep.Name, "error", err)
		result.Error = fmt.Sprintf("OSV check failed: %v", err)
	}

	// Convert OSV vulnerabilities to our format
	for _, osvVuln := range osvVulns {
		vuln := c.convertOSVToVulnerabilityInfo(osvVuln, normalizedDep)
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}

	// Update statistics
	c.updateVulnerabilityStats(result)

	// Generate recommendations
	result.Recommendations = c.generateRecommendations(result)

	slog.Info("Vulnerability check completed",
		"dependency", normalizedDep.Name,
		"vulnerabilities_found", len(result.Vulnerabilities),
		"risk_score", result.RiskScore)

	return result, nil
}

// CheckBatchVulnerabilities checks vulnerabilities for multiple dependencies
func (c *CVEHelper) CheckBatchVulnerabilities(ctx context.Context, dependencies []parser.DependencyInfo) (*BatchVulnerabilityResult, error) {
	result := &BatchVulnerabilityResult{
		Dependencies:           make([]DependencyVulnerabilityResult, 0, len(dependencies)),
		TotalDependencies:      len(dependencies),
		VulnerableDependencies: 0,
		TotalVulnerabilities:   0,
		CheckedAt:              time.Now(),
	}

	slog.Info("Starting batch vulnerability check", "total_dependencies", len(dependencies))

	// Check each dependency
	for i, dep := range dependencies {
		slog.Debug("Checking dependency", "index", i+1, "total", len(dependencies), "name", dep.Name)

		depResult, err := c.CheckDependencyVulnerabilities(ctx, dep)
		if err != nil {
			slog.Warn("Failed to check dependency", "name", dep.Name, "error", err)
			depResult = &DependencyVulnerabilityResult{
				Dependency: dep,
				Error:      err.Error(),
				CheckedAt:  time.Now(),
			}
		}

		result.Dependencies = append(result.Dependencies, *depResult)

		// Update batch statistics
		if depResult.IsVulnerable {
			result.VulnerableDependencies++
		}
		result.TotalVulnerabilities += len(depResult.Vulnerabilities)
	}

	// Calculate overall statistics
	c.updateBatchStats(result)

	slog.Info("Batch vulnerability check completed",
		"total_dependencies", result.TotalDependencies,
		"vulnerable_dependencies", result.VulnerableDependencies,
		"total_vulnerabilities", result.TotalVulnerabilities,
		"overall_risk_score", result.OverallRiskScore)

	return result, nil
}

// checkOSVDatabase queries the OSV (Open Source Vulnerabilities) database
func (c *CVEHelper) checkOSVDatabase(ctx context.Context, dep parser.DependencyInfo) ([]OSVVulnerability, error) {
	ecosystem := c.getEcosystemForRuntime(dep.Runtime)
	if ecosystem == "" {
		return nil, fmt.Errorf("unsupported runtime: %s", dep.Runtime)
	}

	// Ensure the dependency is normalized before querying
	normalizedDep := c.normalizer.NormalizeDependencyInfo(dep)

	// slog.Debug("Querying OSV database",
	// 	"name", normalizedDep.Name,
	// 	"version", normalizedDep.Version,
	// 	"ecosystem", ecosystem,
	// 	"original_name", dep.Name)

	// Prepare query for OSV API
	query := map[string]interface{}{
		"package": map[string]string{
			"name":      normalizedDep.Name,
			"ecosystem": ecosystem,
		},
		"version": normalizedDep.Version,
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.osv.dev/v1/query", strings.NewReader(string(queryBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SilentPatchDetector/1.0")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSV API returned status %d", resp.StatusCode)
	}

	// Parse response
	var osvResp OSVResponse
	if err := json.NewDecoder(resp.Body).Decode(&osvResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return osvResp.Vulns, nil
}

// getEcosystemForRuntime maps runtime types to OSV ecosystems
func (c *CVEHelper) getEcosystemForRuntime(runtime string) string {
	switch strings.ToLower(runtime) {
	case "go":
		return "Go"
	case "node", "npm":
		return "npm"
	case "python", "pip":
		return "PyPI"
	case "java", "maven":
		return "Maven"
	case "gradle":
		return "Maven" // Gradle uses Maven ecosystem
	case "dotnet", "nuget":
		return "NuGet"
	case "ruby", "gem":
		return "RubyGems"
	case "php", "composer":
		return "Packagist"
	case "rust", "cargo":
		return "crates.io"
	default:
		return ""
	}
}

// convertOSVToVulnerabilityInfo converts OSV vulnerability to our format
func (c *CVEHelper) convertOSVToVulnerabilityInfo(osvVuln OSVVulnerability, dep parser.DependencyInfo) VulnerabilityInfo {
	vuln := VulnerabilityInfo{
		ID:               osvVuln.ID,
		Summary:          osvVuln.Summary,
		Description:      osvVuln.Details,
		PublishedDate:    time.Now(), // Default to current time since not available in existing structure
		ModifiedDate:     time.Now(), // Default to current time since not available in existing structure
		AffectedVersions: []string{},
		PatchedVersions:  []string{},
		References:       []string{},
		Severity:         SeverityMedium, // Default severity since not available in existing structure
		Score:            5.0,            // Default score
	}

	// Extract CVE ID from ID if it contains CVE
	if strings.Contains(strings.ToUpper(osvVuln.ID), "CVE-") {
		vuln.CVE = osvVuln.ID
	}

	// Extract affected and patched versions from existing structure
	for _, affected := range osvVuln.Affects {
		for _, r := range affected.Ranges {
			for _, event := range r.Events {
				if event.Introduced != "" {
					vuln.AffectedVersions = append(vuln.AffectedVersions, event.Introduced)
				}
				if event.Fixed != "" {
					vuln.PatchedVersions = append(vuln.PatchedVersions, event.Fixed)
				}
			}
		}
	}

	// Extract references
	for _, ref := range osvVuln.References {
		vuln.References = append(vuln.References, ref.URL)
	}

	return vuln
}

// updateVulnerabilityStats updates statistics for a dependency result
func (c *CVEHelper) updateVulnerabilityStats(result *DependencyVulnerabilityResult) {
	result.TotalCount = len(result.Vulnerabilities)
	result.IsVulnerable = result.TotalCount > 0

	totalScore := 0.0
	for _, vuln := range result.Vulnerabilities {
		switch vuln.Severity {
		case SeverityCritical:
			result.CriticalCount++
		case SeverityHigh:
			result.HighCount++
		case SeverityMedium:
			result.MediumCount++
		case SeverityLow:
			result.LowCount++
		}
		totalScore += vuln.Score
	}

	if result.TotalCount > 0 {
		result.RiskScore = totalScore / float64(result.TotalCount)
	}
}

// updateBatchStats calculates overall statistics for batch results
func (c *CVEHelper) updateBatchStats(result *BatchVulnerabilityResult) {
	totalScore := 0.0
	totalVulns := 0
	highestSeverity := SeverityInfo

	for _, dep := range result.Dependencies {
		result.Summary.CriticalCount += dep.CriticalCount
		result.Summary.HighCount += dep.HighCount
		result.Summary.MediumCount += dep.MediumCount
		result.Summary.LowCount += dep.LowCount

		totalScore += dep.RiskScore * float64(dep.TotalCount)
		totalVulns += dep.TotalCount

		// Update highest severity
		for _, vuln := range dep.Vulnerabilities {
			if c.severityPriority(vuln.Severity) > c.severityPriority(highestSeverity) {
				highestSeverity = vuln.Severity
			}
		}
	}

	if totalVulns > 0 {
		result.OverallRiskScore = totalScore / float64(totalVulns)
	}
	result.HighestSeverity = highestSeverity
}

// severityPriority returns numeric priority for severity comparison
func (c *CVEHelper) severityPriority(severity CVESeverity) int {
	switch severity {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

// generateRecommendations generates security recommendations based on vulnerabilities
func (c *CVEHelper) generateRecommendations(result *DependencyVulnerabilityResult) []string {
	recommendations := []string{}

	if !result.IsVulnerable {
		recommendations = append(recommendations, "No known vulnerabilities found for this dependency.")
		return recommendations
	}

	if result.CriticalCount > 0 || result.HighCount > 0 {
		recommendations = append(recommendations, "URGENT: Update this dependency immediately due to critical/high severity vulnerabilities.")
	}

	// Suggest version updates
	if len(result.Vulnerabilities) > 0 {
		recommendations = append(recommendations, "Review patched versions and update to the latest secure version.")
	}

	if result.RiskScore > 7.0 {
		recommendations = append(recommendations, "Consider finding alternative dependencies with better security track records.")
	}

	// Add specific recommendations based on vulnerability types
	for _, vuln := range result.Vulnerabilities {
		if vuln.AttackVector == "NETWORK" {
			recommendations = append(recommendations, "Network-based vulnerabilities detected. Review network exposure and access controls.")
			break
		}
	}

	return recommendations
}

// GetVulnerabilityByID retrieves detailed information about a specific vulnerability
func (c *CVEHelper) GetVulnerabilityByID(ctx context.Context, vulnID string) (*VulnerabilityInfo, error) {
	encodedID := url.QueryEscape(vulnID)
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.osv.dev/v1/vulns/%s", encodedID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "SilentPatchDetector/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("vulnerability not found: %s", vulnID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSV API returned status %d", resp.StatusCode)
	}

	var osvVuln OSVVulnerability
	if err := json.NewDecoder(resp.Body).Decode(&osvVuln); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our format (using empty dependency as we don't have context)
	vuln := c.convertOSVToVulnerabilityInfo(osvVuln, parser.DependencyInfo{})
	return &vuln, nil
}

// FilterVulnerabilitiesBySeverity filters vulnerabilities by minimum severity level
func (c *CVEHelper) FilterVulnerabilitiesBySeverity(result *BatchVulnerabilityResult, minSeverity CVESeverity) *BatchVulnerabilityResult {
	minPriority := c.severityPriority(minSeverity)

	filtered := &BatchVulnerabilityResult{
		Dependencies:      make([]DependencyVulnerabilityResult, 0),
		TotalDependencies: result.TotalDependencies,
		CheckedAt:         result.CheckedAt,
	}

	for _, dep := range result.Dependencies {
		filteredDep := DependencyVulnerabilityResult{
			Dependency:      dep.Dependency,
			Vulnerabilities: make([]VulnerabilityInfo, 0),
			CheckedAt:       dep.CheckedAt,
			Error:           dep.Error,
		}

		for _, vuln := range dep.Vulnerabilities {
			if c.severityPriority(vuln.Severity) >= minPriority {
				filteredDep.Vulnerabilities = append(filteredDep.Vulnerabilities, vuln)
			}
		}

		c.updateVulnerabilityStats(&filteredDep)
		filtered.Dependencies = append(filtered.Dependencies, filteredDep)
	}

	c.updateBatchStats(filtered)
	return filtered
}

// AggregateVulnerabilitySummary calculates the summary from findings
func AggregateVulnerabilitySummary(findings []model.ScanFinding) model.ScanSummary {
	severityCount := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"none":     0,
		"ignored":  0,
	}
	totalVulns := 0

	for _, f := range findings {
		sev := strings.ToLower(f.Severity)

		vulnCount := len(f.VulnerabilityIDs)
		totalVulns += vulnCount

		if vulnCount == 0 {
			severityCount["none"]++
		} else if sev == "critical" || sev == "high" || sev == "medium" || sev == "low" {
			severityCount[sev]++
		} else {
			severityCount["ignored"]++
		}
	}

	return model.ScanSummary{
		TotalDependencies:    len(findings),
		TotalVulnerabilities: totalVulns,
		Critical:             severityCount["critical"],
		High:                 severityCount["high"],
		Medium:               severityCount["medium"],
		Low:                  severityCount["low"],
		Ignored:              severityCount["ignored"],
		None:                 severityCount["none"],
	}
}

// EvaluatePolicy determines fail/pass status based on summary and policy
func EvaluatePolicy(summary model.ScanSummary, failOn []string) (status, reason string) {
	for _, sev := range failOn {
		switch sev {
		case "critical":
			if summary.Critical > 0 {
				return "fail", "Critical severity vulnerabilities found"
			}
		case "high":
			if summary.High > 0 {
				return "fail", "High severity vulnerabilities found"
			}
		}
	}
	return "pass", "No blocking vulnerabilities found"
}
