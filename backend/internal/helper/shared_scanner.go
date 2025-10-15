package helper

import (
	"context"

	"elang-backend/internal/helper/parser"
	"elang-backend/internal/model"
	"fmt"
	"log/slog"

	"sync"
)

// SharedScanner provides reusable scanning functionality across services
type SharedScanner struct {
	cveService    *CVEHelper
	maxConcurrent int
}

// NewSharedScanner creates a new shared scanner with controlled concurrency
func NewSharedScanner(maxConcurrent int) *SharedScanner {
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // default
	}
	return &SharedScanner{
		cveService:    NewCVEHelper(),
		maxConcurrent: maxConcurrent,
	}
}

// ScanDependenciesWithControl scans dependencies with controlled concurrency using semaphore pattern
func (ss *SharedScanner) ScanDependenciesWithControl(
	ctx context.Context,
	dependencies []DependencyInfo,
) (findings []model.ScanFinding, depsWithVulns []DependencyWithVulnerabilities, totalCritical, totalHigh, totalMedium, totalLow int) {

	if len(dependencies) == 0 {
		return
	}

	var (
		mu        sync.Mutex
		wg        sync.WaitGroup
		semaphore = make(chan struct{}, ss.maxConcurrent)
	)

	findings = make([]model.ScanFinding, 0)
	depsWithVulns = make([]DependencyWithVulnerabilities, 0)

	// Process each dependency with controlled concurrency
	for i, dep := range dependencies {
		wg.Add(1)

		// Acquire semaphore slot
		semaphore <- struct{}{}

		go func(dependency parser.DependencyInfo, index int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Check for context cancellation
			select {
			case <-ctx.Done():
				slog.Warn("Scan cancelled", "dependency", dependency.Name)
				return
			default:
			}

			// Perform vulnerability check
			result, err := ss.cveService.CheckDependencyVulnerabilities(ctx, dependency)
			if err != nil {
				slog.Warn("Failed to check vulnerabilities", "dependency", dependency.Name, "error", err)
				return
			}

			// Determine severity
			severity := "none"
			if result.CriticalCount > 0 {
				severity = "critical"
			} else if result.HighCount > 0 {
				severity = "high"
			} else if result.MediumCount > 0 {
				severity = "medium"
			} else if result.LowCount > 0 {
				severity = "low"
			}

			// Extract vulnerability IDs
			var vulnIDs []string
			for _, v := range result.Vulnerabilities {
				vulnIDs = append(vulnIDs, v.ID)
			}

			// Get recommendation
			recommendation := ""
			if len(result.Recommendations) > 0 {
				recommendation = result.Recommendations[0]
			}

			// Create finding
			finding := model.ScanFinding{
				Dependency:       dependency.Name,
				Version:          dependency.Version,
				Severity:         severity,
				VulnerabilityIDs: vulnIDs,
				Recommendation:   recommendation,
			}

			// Create enhanced dependency with vulnerabilities
			depWithVuln := DependencyWithVulnerabilities{
				Name:            dependency.Name,
				Version:         dependency.Version,
				Owner:           dependency.Owner,
				Repo:            dependency.Repo,
				RepositoryURL:   dependency.GitHubURL,
				Runtime:         dependency.Runtime,
				IsGitHub:        dependency.IsGitHubRepo,
				Vulnerabilities: result.Vulnerabilities,
				RiskScore:       result.RiskScore,
			}

			// Update results (thread-safe)
			mu.Lock()
			findings = append(findings, finding)
			depsWithVulns = append(depsWithVulns, depWithVuln)
			totalCritical += result.CriticalCount
			totalHigh += result.HighCount
			totalMedium += result.MediumCount
			totalLow += result.LowCount
			mu.Unlock()

			slog.Debug("Dependency scanned",
				"dependency", dependency.Name,
				"vulnerabilities", len(result.Vulnerabilities),
				"severity", severity,
				"progress", fmt.Sprintf("%d/%d", index+1, len(dependencies)))
		}(dep, i)
	}

	// Wait for all scans to complete
	wg.Wait()
	close(semaphore)

	slog.Info("Dependency scan completed",
		"total", len(dependencies),
		"critical", totalCritical,
		"high", totalHigh,
		"medium", totalMedium,
		"low", totalLow)

	return
}
