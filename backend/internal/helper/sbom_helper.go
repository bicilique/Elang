package helper

import (
	"elang-backend/internal/model"
	"encoding/json"
	"fmt"

	"strings"
	"time"

	"github.com/google/uuid"
)

// CycloneDXSBOM represents the SBOM structure
// Only essential fields for CI/CD and future storage
// License, purl, etc. are placeholders for now

type CycloneDXSBOM struct {
	BomFormat       string                    `json:"bomFormat"`
	SpecVersion     string                    `json:"specVersion"`
	Version         int                       `json:"version"`
	SerialNumber    string                    `json:"serialNumber"`
	Metadata        CycloneDXMetadata         `json:"metadata"`
	Components      []CycloneDXComponent      `json:"components"`
	Vulnerabilities []CycloneDXVulnerability  `json:"vulnerabilities,omitempty"`
	Dependencies    []CycloneDXDependencyNode `json:"dependencies,omitempty"`
}

type CycloneDXMetadata struct {
	Timestamp  string                 `json:"timestamp"`
	Tools      []CycloneDXTool        `json:"tools"`
	Component  CycloneDXComponentMeta `json:"component"`
	Properties []CycloneDXProperty    `json:"properties,omitempty"`
}

type CycloneDXProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CycloneDXTool struct {
	Vendor  string `json:"vendor"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type CycloneDXComponentMeta struct {
	Type       string              `json:"type"`
	Name       string              `json:"name"`
	Version    string              `json:"version"`
	Properties []CycloneDXProperty `json:"properties,omitempty"`
}

type CycloneDXComponent struct {
	BomRef       string                   `json:"bom-ref"`
	Type         string                   `json:"type"`
	Group        string                   `json:"group,omitempty"`
	Name         string                   `json:"name"`
	Version      string                   `json:"version"`
	Purl         string                   `json:"purl,omitempty"`
	Licenses     []CycloneDXLicenseHolder `json:"licenses,omitempty"`
	Hashes       []CycloneDXHash          `json:"hashes,omitempty"`
	ExternalRefs []CycloneDXExternalRef   `json:"externalReferences,omitempty"`
	Properties   []CycloneDXProperty      `json:"properties,omitempty"`
}

type CycloneDXHash struct {
	Algorithm string `json:"alg"`
	Content   string `json:"content"`
}

type CycloneDXExternalRef struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type CycloneDXLicenseHolder struct {
	License CycloneDXLicense `json:"license"`
}

type CycloneDXLicense struct {
	ID string `json:"id"`
}

type CycloneDXVulnerability struct {
	BomRef      string                       `json:"bom-ref"`
	ID          string                       `json:"id"`
	Source      CycloneDXVulnerabilitySource `json:"source"`
	Description string                       `json:"description,omitempty"`
	Detail      string                       `json:"detail,omitempty"`
	Ratings     []CycloneDXRating            `json:"ratings,omitempty"`
	Cwes        []int                        `json:"cwes,omitempty"`
	Advisories  []CycloneDXAdvisory          `json:"advisories,omitempty"`
	Published   string                       `json:"published,omitempty"`
	Updated     string                       `json:"updated,omitempty"`
	Affects     []CycloneDXAffect            `json:"affects,omitempty"`
}

type CycloneDXVulnerabilitySource struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type CycloneDXRating struct {
	Source   CycloneDXVulnerabilitySource `json:"source,omitempty"`
	Score    float64                      `json:"score,omitempty"`
	Severity string                       `json:"severity,omitempty"`
	Method   string                       `json:"method,omitempty"`
	Vector   string                       `json:"vector,omitempty"`
}

type CycloneDXAdvisory struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url"`
}

type CycloneDXAffect struct {
	Ref      string                  `json:"ref"`
	Versions []CycloneDXVersionRange `json:"versions,omitempty"`
}

type CycloneDXVersionRange struct {
	Version string `json:"version,omitempty"`
	Range   string `json:"range,omitempty"`
	Status  string `json:"status,omitempty"`
}

type CycloneDXDependencyNode struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

// EnhancedSBOMData contains all information needed to generate comprehensive SBOM
type EnhancedSBOMData struct {
	AppID         string
	AppName       string
	AppVersion    string
	Runtime       string
	Framework     string
	Dependencies  []DependencyWithVulnerabilities
	ScanTimestamp time.Time
	TotalFindings int
	CriticalCount int
	HighCount     int
	MediumCount   int
	LowCount      int
}

// DependencyWithVulnerabilities contains dependency info with its vulnerabilities
type DependencyWithVulnerabilities struct {
	Name            string
	Version         string
	Owner           string
	Repo            string
	RepositoryURL   string
	Runtime         string
	IsGitHub        bool
	Vulnerabilities []VulnerabilityInfo
	RiskScore       float64
}

// GenerateEnhancedCycloneDXSBOM generates a comprehensive CycloneDX SBOM with vulnerability data
func GenerateEnhancedCycloneDXSBOM(data EnhancedSBOMData) ([]byte, error) {
	timestamp := data.ScanTimestamp
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	bom := CycloneDXSBOM{
		BomFormat:    "CycloneDX",
		SpecVersion:  "1.5",
		Version:      1,
		SerialNumber: "urn:uuid:" + uuid.New().String(),
		Metadata: CycloneDXMetadata{
			Timestamp: timestamp.Format(time.RFC3339),
			Tools: []CycloneDXTool{{
				Vendor:  "Silent Patch Detector",
				Name:    "dependency-vulnerability-scanner",
				Version: "2.0.0",
			}},
			Component: CycloneDXComponentMeta{
				Type:    "application",
				Name:    data.AppName,
				Version: data.AppVersion,
				Properties: []CycloneDXProperty{
					{Name: "app:id", Value: data.AppID},
					{Name: "app:runtime", Value: data.Runtime},
					{Name: "app:framework", Value: data.Framework},
					{Name: "scan:timestamp", Value: timestamp.Format(time.RFC3339)},
					{Name: "scan:total_findings", Value: fmt.Sprintf("%d", data.TotalFindings)},
					{Name: "scan:critical_count", Value: fmt.Sprintf("%d", data.CriticalCount)},
					{Name: "scan:high_count", Value: fmt.Sprintf("%d", data.HighCount)},
					{Name: "scan:medium_count", Value: fmt.Sprintf("%d", data.MediumCount)},
					{Name: "scan:low_count", Value: fmt.Sprintf("%d", data.LowCount)},
				},
			},
		},
		Components:      []CycloneDXComponent{},
		Vulnerabilities: []CycloneDXVulnerability{},
		Dependencies:    []CycloneDXDependencyNode{},
	}

	// Track component refs for dependency graph
	componentRefs := make(map[string]bool)

	// Process each dependency
	for _, dep := range data.Dependencies {
		bomRef := generateBomRef(dep.Name, dep.Version)
		componentRefs[bomRef] = true

		// Determine package URL (purl) based on runtime
		purl := generatePurl(dep.Runtime, dep.Owner, dep.Repo, dep.Name, dep.Version)

		// Build external references
		var externalRefs []CycloneDXExternalRef
		if dep.RepositoryURL != "" {
			externalRefs = append(externalRefs, CycloneDXExternalRef{
				Type: "vcs",
				URL:  dep.RepositoryURL,
			})
		}

		// Build component properties
		properties := []CycloneDXProperty{
			{Name: "dependency:owner", Value: dep.Owner},
			{Name: "dependency:repo", Value: dep.Repo},
			{Name: "dependency:runtime", Value: dep.Runtime},
			{Name: "dependency:is_github", Value: fmt.Sprintf("%t", dep.IsGitHub)},
			{Name: "dependency:risk_score", Value: fmt.Sprintf("%.2f", dep.RiskScore)},
			{Name: "dependency:vulnerability_count", Value: fmt.Sprintf("%d", len(dep.Vulnerabilities))},
		}

		component := CycloneDXComponent{
			BomRef:       bomRef,
			Type:         "library",
			Group:        dep.Owner,
			Name:         dep.Name,
			Version:      dep.Version,
			Purl:         purl,
			ExternalRefs: externalRefs,
			Properties:   properties,
		}

		bom.Components = append(bom.Components, component)

		// Process vulnerabilities for this component
		for _, vuln := range dep.Vulnerabilities {
			vulnBomRef := generateVulnBomRef(vuln.ID, bomRef)

			// Build ratings
			var ratings []CycloneDXRating
			if vuln.Score > 0 {
				ratings = append(ratings, CycloneDXRating{
					Source: CycloneDXVulnerabilitySource{
						Name: "NVD",
						URL:  "https://nvd.nist.gov/",
					},
					Score:    vuln.Score,
					Severity: string(vuln.Severity),
					Method:   "CVSSv3",
					Vector:   vuln.VectorString,
				})
			}

			// Build advisories from references
			var advisories []CycloneDXAdvisory
			for _, ref := range vuln.References {
				advisories = append(advisories, CycloneDXAdvisory{
					URL: ref,
				})
			}

			// Build affected versions
			var affects []CycloneDXAffect
			if len(vuln.AffectedVersions) > 0 {
				var versionRanges []CycloneDXVersionRange
				for _, affectedVer := range vuln.AffectedVersions {
					versionRanges = append(versionRanges, CycloneDXVersionRange{
						Version: affectedVer,
						Status:  "affected",
					})
				}
				affects = append(affects, CycloneDXAffect{
					Ref:      bomRef,
					Versions: versionRanges,
				})
			}

			published := ""
			if !vuln.PublishedDate.IsZero() {
				published = vuln.PublishedDate.Format(time.RFC3339)
			}
			updated := ""
			if !vuln.ModifiedDate.IsZero() {
				updated = vuln.ModifiedDate.Format(time.RFC3339)
			}

			cycloneDXVuln := CycloneDXVulnerability{
				BomRef: vulnBomRef,
				ID:     vuln.CVE,
				Source: CycloneDXVulnerabilitySource{
					Name: "OSV",
					URL:  "https://osv.dev/",
				},
				Description: vuln.Summary,
				Detail:      vuln.Description,
				Ratings:     ratings,
				Advisories:  advisories,
				Published:   published,
				Updated:     updated,
				Affects:     affects,
			}

			bom.Vulnerabilities = append(bom.Vulnerabilities, cycloneDXVuln)
		}
	}

	// Build dependency graph (simplified - application depends on all components)
	appRef := "app:" + data.AppName
	var dependsOn []string
	for ref := range componentRefs {
		dependsOn = append(dependsOn, ref)
	}
	bom.Dependencies = append(bom.Dependencies, CycloneDXDependencyNode{
		Ref:       appRef,
		DependsOn: dependsOn,
	})

	return json.MarshalIndent(bom, "", "  ")
}

// GenerateCycloneDXSBOM generates a CycloneDX SBOM from scan result (legacy support)
func GenerateCycloneDXSBOM(scanResult model.ScanApplicationResult) ([]byte, error) {
	// Convert ScanApplicationResult to EnhancedSBOMData
	data := EnhancedSBOMData{
		AppID:         scanResult.AppID,
		AppName:       scanResult.AppName,
		AppVersion:    "1.0.0", // default
		ScanTimestamp: time.Now().UTC(),
		TotalFindings: len(scanResult.Findings),
		CriticalCount: scanResult.Summary.Critical,
		HighCount:     scanResult.Summary.High,
		MediumCount:   scanResult.Summary.Medium,
		LowCount:      scanResult.Summary.Low,
	}

	// Convert findings to dependencies
	for _, finding := range scanResult.Findings {
		// Parse dependency name
		parts := strings.Split(finding.Dependency, ":")
		name := finding.Dependency
		group := ""
		if len(parts) == 2 {
			group = parts[0]
			name = parts[1]
		}

		dep := DependencyWithVulnerabilities{
			Name:    name,
			Owner:   group,
			Version: finding.Version,
		}

		// Create vulnerabilities from finding
		for _, vulnID := range finding.VulnerabilityIDs {
			vuln := VulnerabilityInfo{
				ID:       vulnID,
				CVE:      vulnID,
				Summary:  finding.Recommendation,
				Severity: CVESeverity(strings.ToUpper(finding.Severity)),
			}
			dep.Vulnerabilities = append(dep.Vulnerabilities, vuln)
		}

		data.Dependencies = append(data.Dependencies, dep)
	}

	return GenerateEnhancedCycloneDXSBOM(data)
}

// generateBomRef creates a unique BOM reference for a component
func generateBomRef(name, version string) string {
	return fmt.Sprintf("pkg:%s@%s", name, version)
}

// generateVulnBomRef creates a unique BOM reference for a vulnerability
func generateVulnBomRef(vulnID, componentRef string) string {
	return fmt.Sprintf("vuln:%s:%s", vulnID, componentRef)
}

// generatePurl generates a package URL based on runtime/ecosystem
func generatePurl(runtime, owner, repo, name, version string) string {
	runtimeLower := strings.ToLower(runtime)

	switch {
	case strings.Contains(runtimeLower, "java") || strings.Contains(runtimeLower, "maven"):
		if owner != "" {
			return fmt.Sprintf("pkg:maven/%s/%s@%s", owner, name, version)
		}
		return fmt.Sprintf("pkg:maven/%s@%s", name, version)

	case strings.Contains(runtimeLower, "node") || strings.Contains(runtimeLower, "npm"):
		if owner != "" {
			return fmt.Sprintf("pkg:npm/%s/%s@%s", owner, name, version)
		}
		return fmt.Sprintf("pkg:npm/%s@%s", name, version)

	case strings.Contains(runtimeLower, "python") || strings.Contains(runtimeLower, "pip"):
		return fmt.Sprintf("pkg:pypi/%s@%s", name, version)

	case strings.Contains(runtimeLower, "go") || strings.Contains(runtimeLower, "golang"):
		if repo != "" && owner != "" {
			return fmt.Sprintf("pkg:golang/%s/%s@%s", owner, repo, version)
		}
		return fmt.Sprintf("pkg:golang/%s@%s", name, version)

	case strings.Contains(runtimeLower, "ruby") || strings.Contains(runtimeLower, "gem"):
		return fmt.Sprintf("pkg:gem/%s@%s", name, version)

	case strings.Contains(runtimeLower, "rust") || strings.Contains(runtimeLower, "cargo"):
		return fmt.Sprintf("pkg:cargo/%s@%s", name, version)

	case strings.Contains(runtimeLower, "php") || strings.Contains(runtimeLower, "composer"):
		if owner != "" {
			return fmt.Sprintf("pkg:composer/%s/%s@%s", owner, name, version)
		}
		return fmt.Sprintf("pkg:composer/%s@%s", name, version)

	case strings.Contains(runtimeLower, "nuget") || strings.Contains(runtimeLower, ".net"):
		return fmt.Sprintf("pkg:nuget/%s@%s", name, version)

	default:
		// Generic package URL
		return fmt.Sprintf("pkg:generic/%s@%s", name, version)
	}
}

// findColon returns the index of the first colon, or -1
func findColon(s string) int {
	for i, c := range s {
		if c == ':' {
			return i
		}
	}
	return -1
}
