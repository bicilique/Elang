package services

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/helper"
	"elang-backend/internal/helper/parser"
	"elang-backend/internal/model"
	"elang-backend/internal/model/dto"
	"elang-backend/internal/repository"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationService struct {
	depedencyParserService helper.DependencyParser
	cveService             *helper.CVEHelper
	githubApiService       GitHubAPIInterface
	objectStorageService   ObjectStorageInterface

	// Add fields as necessary, e.g., database connection, logger, etc.
	appRepository              repository.ApplicationRepository
	depedencyRepository        repository.DependencyRepository
	appToDepedencyRepository   repository.AppDependencyRepository
	depedencyVersionRepository repository.DependencyVersionRepository
	runTimeRepository          repository.RuntimeRepository
	frameWorkRepository        repository.FrameworkRepository
	auditTrailRepository       repository.AuditTrailRepository
}

func NewApplicationService(basicRepo dto.BasicRepositories,
	githubService GitHubAPIInterface,
	dependencyParser helper.DependencyParser,
	objectStorageService ObjectStorageInterface,
) ApplicationInterface {
	return &ApplicationService{
		githubApiService:       githubService,
		objectStorageService:   objectStorageService,
		depedencyParserService: dependencyParser,
		cveService:             helper.NewCVEHelper(),

		appRepository:              basicRepo.AppRepository,
		depedencyRepository:        basicRepo.DepedencyRepository,
		appToDepedencyRepository:   basicRepo.AppToDepedencyRepository,
		depedencyVersionRepository: basicRepo.DepedencyVersionRepository,
		runTimeRepository:          basicRepo.RunTimeRepository,
		frameWorkRepository:        basicRepo.FrameWorkRepository,
		auditTrailRepository:       basicRepo.AuditTrailRepository,
	}
}

func (m *ApplicationService) AddApplication(ctx context.Context, appName, runtimeType, framework, description, fileName, content string) (*model.AddApplicationResponse, error) {
	// Check for empty inputs
	if content == "" || fileName == "" || runtimeType == "" || appName == "" {
		return nil, fmt.Errorf("content, file name, runtime type, and application name cannot be empty")
	}

	// Check for valid runtime type (case-insensitive)
	runtime, err := m.runTimeRepository.GetByNameCI(ctx, runtimeType)
	if err != nil {
		return nil, err
	}
	if runtime == nil {
		return nil, fmt.Errorf("runtime type %s not found", runtimeType)
	}

	// Check for valid framework (case-insensitive)
	frameworkEntity, err := m.frameWorkRepository.GetByNameCI(ctx, framework)
	if err != nil {
		return nil, err
	}
	if frameworkEntity == nil {
		return nil, fmt.Errorf("framework %s not found for runtime %s", framework, runtimeType)
	}

	// Check if app already exists
	app, err := m.appRepository.GetByName(ctx, appName)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if app != nil {
		return nil, fmt.Errorf("application with name %s already exists", appName)
	}

	// Create and save new application
	newApp := &entity.App{
		ID:          uuid.New(),
		Name:        appName,
		RuntimeID:   &runtime.ID,
		FrameworkID: &frameworkEntity.ID,
		Description: &description,
		Status:      "inactive",
	}
	if err := m.appRepository.Create(ctx, newApp); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Audit trail: Application created
	err = m.auditApplicationAction(ctx, newApp.ID, "application_created", nil, map[string]interface{}{
		"app_name":     appName,
		"runtime_type": runtimeType,
		"framework":    framework,
		"description":  description,
		"file_name":    fileName,
	})
	if err != nil {
		slog.Warn("Failed to create audit trail for application creation", "error", err)
	}

	// Dependencies: process in background
	deps := m.depedencyParserService.ParseDependencyFileWithGitHub(fileName, content, helper.GetRuntimeTypeCI(runtimeType))
	go func() {
		bgCtx := context.Background()
		var (
			wg    sync.WaitGroup
			errCh = make(chan error, len(deps.Dependencies))
		)
		for _, dep := range deps.Dependencies {
			wg.Add(1)
			depCopy := dep
			go func(dep helper.DependencyInfo) {
				defer wg.Done()
				m.processDependency(bgCtx, dep, newApp, errCh)
			}(depCopy)
		}
		wg.Wait()
		close(errCh)
		var depErrors []string
		for err := range errCh {
			if err != nil {
				depErrors = append(depErrors, err.Error())
			}
		}
		// Update app status after processing
		finalStatus := "active"
		if len(depErrors) > 0 {
			finalStatus = "inactive"
		}
		// Only update the status field to avoid overwriting other fields
		if err := m.appRepository.UpdateStatus(bgCtx, newApp.ID, finalStatus); err != nil {
			slog.Error("failed to update app status after dependency processing", "error", err)
		}
	}()

	message := "Application created, dependency processing started in background."
	response := &model.AddApplicationResponse{
		AppID:           fmt.Sprintf("%v", newApp.ID),
		AppName:         newApp.Name,
		RuntimeType:     runtimeType,
		Framework:       framework,
		Description:     description,
		Status:          newApp.Status,
		DependencyParse: deps.Dependencies,
		Message:         message,
	}

	return response, nil
}

func (m *ApplicationService) AddApplicationDependency(ctx context.Context, appUID string, deps []model.DependencyInfoRequest) (interface{}, error) {
	// Parse app UUID
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	// Check if app exists
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	results := make(map[string]interface{})
	successful, failed := 0, 0

	for _, depInfo := range deps {
		// Validate GitHub repo info if flagged
		slog.Info("Adding dependency", "name", depInfo.Name, "owner", depInfo.Owner, "repo", depInfo.Repo, "version", depInfo.Version, "is_github", depInfo.IsGitHubRepo)
		if depInfo.IsGitHubRepo {
			owner, repo, valid := depInfo.Owner, depInfo.Repo, false
			if depInfo.RepositoryURL != "" {
				parts, isValid := helper.ExtractGitHubOwnerRepo(depInfo.RepositoryURL)
				if isValid {
					owner, repo, valid = parts.Owner, parts.Repo, true
				}
			}
			if !valid && owner != "" && repo != "" {
				valid = true
			}
			if valid {
				repoInfo, err := m.githubApiService.GetRepoInfo(owner, repo)
				if err == nil && repoInfo != nil {
					depInfo.Owner, depInfo.Repo = owner, repo
					depInfo.RepositoryURL = fmt.Sprintf("https://github.com/%s/%s", owner, repo)
					depInfo.IsGitHubRepo = true
				} else {
					depInfo.IsGitHubRepo = false
					depInfo.RepositoryURL = ""
					slog.Warn("Invalid GitHub repo, marking as non-GitHub", "owner", owner, "repo", repo)
				}
			} else {
				depInfo.IsGitHubRepo = false
				depInfo.RepositoryURL = ""
				slog.Warn("No valid GitHub info, marking as non-GitHub repo")
			}
		}

		// Lookup dependency
		dependency, err := m.depedencyRepository.GetByOwnerRepo(ctx, depInfo.Owner, depInfo.Repo)
		if err != nil && err != gorm.ErrRecordNotFound {
			results[fmt.Sprintf("%s/%s", depInfo.Owner, depInfo.Repo)] = map[string]interface{}{
				"status": "failed", "error": fmt.Sprintf("database error: %v", err),
			}
			failed++
			continue
		}

		slog.Info("Processing dependency", "name", depInfo.Name, "owner", depInfo.Owner, "repo", depInfo.Repo, "version", depInfo.Version, "is_github", depInfo.IsGitHubRepo)
		// Get default branch if GitHub repo
		var defaultBranch string
		if depInfo.IsGitHubRepo {
			defaultBranch, _ = m.githubApiService.GetDefaultBranch(depInfo.Owner, depInfo.Repo)
		}

		// Create dependency if not found
		if dependency == nil {
			dependency = &entity.Dependency{
				ID:            uuid.New(),
				Name:          depInfo.Name,
				Owner:         depInfo.Owner,
				Repo:          depInfo.Repo,
				DefaultBranch: &defaultBranch,
				RepositoryURL: &depInfo.RepositoryURL,
			}
			if err := m.depedencyRepository.Create(ctx, dependency); err != nil {
				results[fmt.Sprintf("%s/%s", depInfo.Owner, depInfo.Repo)] = map[string]interface{}{
					"status": "failed", "error": fmt.Sprintf("failed to create dependency: %v", err),
				}
				failed++
				continue
			}
			// Re-fetch to ensure latest state
			dependency, _ = m.depedencyRepository.GetByOwnerRepo(ctx, depInfo.Owner, depInfo.Repo)
		} else if depInfo.IsGitHubRepo && (dependency.DefaultBranch == nil || *dependency.DefaultBranch == "" || dependency.RepositoryURL == nil) {
			// Update missing fields
			if defaultBranch != "" {
				dependency.DefaultBranch = &defaultBranch
			}
			dependency.RepositoryURL = &depInfo.RepositoryURL
			dependency.Owner = depInfo.Owner
			dependency.Repo = depInfo.Repo
			if err := m.depedencyRepository.Update(ctx, dependency); err != nil {
				slog.Warn("failed to update dependency default branch", "error", err, "dependency_id", dependency.ID)
			}
			// Re-fetch to ensure latest state
			dependency, _ = m.depedencyRepository.GetByOwnerRepo(ctx, depInfo.Owner, depInfo.Repo)
		}

		// Check if app-dependency relationship already exists
		existingAppDep, err := m.appToDepedencyRepository.GetByAppAndDependencyID(ctx, appID, dependency.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			results[fmt.Sprintf("%s/%s", depInfo.Owner, depInfo.Repo)] = map[string]interface{}{
				"status": "failed", "error": fmt.Sprintf("database error checking app dependency: %v", err),
			}
			failed++
			continue
		}
		if existingAppDep != nil {
			results[fmt.Sprintf("%s/%s", depInfo.Owner, depInfo.Repo)] = map[string]interface{}{
				"status": "skipped", "reason": "dependency already exists for this application",
			}
			continue
		}

		// Find matching tag if GitHub repo
		if depInfo.IsGitHubRepo {
			if matchedVersion, err := m.githubApiService.FindMatchingTag(depInfo.Owner, depInfo.Repo, depInfo.Version); err == nil && matchedVersion != "" {
				depInfo.Version = matchedVersion
			}
		}

		appDependency := &entity.AppDependency{
			ID:           uuid.New(),
			AppID:        appID,
			DependencyID: dependency.ID,
			UsedVersion:  depInfo.Version,
			IsMonitored:  false,
		}
		if err := m.appToDepedencyRepository.Create(ctx, appDependency); err != nil {
			results[fmt.Sprintf("%s/%s", depInfo.Owner, depInfo.Repo)] = map[string]interface{}{
				"status": "failed", "error": fmt.Sprintf("failed to create app dependency: %v", err),
			}
			failed++
			continue
		}
		results[fmt.Sprintf("%s/%s", depInfo.Owner, depInfo.Repo)] = map[string]interface{}{
			"status":        "success",
			"dependency_id": dependency.ID.String(),
			"app_dep_id":    appDependency.ID.String(),
		}
		successful++
	}

	return map[string]interface{}{
		"app_id":   app.ID.String(),
		"app_name": app.Name,
		"results":  results,
		"summary": map[string]interface{}{
			"total":      len(deps),
			"successful": successful,
			"failed":     failed,
		},
	}, nil
}

func (m *ApplicationService) ListApplicationDependency(ctx context.Context, appUID string) (*model.ListApplicationDependencyResponse, error) {
	// Find the app by ID (UUID)
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app: %w", err)
	}
	if app == nil {
		return nil, fmt.Errorf("application not found")
	}

	// Get all app dependencies for this app
	appDeps, err := m.appToDepedencyRepository.GetByAppID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app dependencies: %w", err)
	}

	var depDetails []model.ApplicationDependencyDetail
	for _, appDep := range appDeps {
		dep, err := m.depedencyRepository.GetByID(ctx, appDep.DependencyID)
		if err != nil || dep == nil {
			continue // skip missing dependency
		}

		depDetails = append(depDetails, model.ApplicationDependencyDetail{
			DependencyID:  dep.ID.String(),
			Name:          dep.Name,
			Owner:         dep.Owner,
			Repo:          dep.Repo,
			UsedVersion:   appDep.UsedVersion,
			IsMonitored:   appDep.IsMonitored,
			RepositoryURL: derefString(dep.RepositoryURL),
			LastTag:       dep.LastTag,
			DefaultBranch: dep.DefaultBranch,
		})
	}

	return &model.ListApplicationDependencyResponse{
		AppID:        app.ID.String(),
		AppName:      app.Name,
		Dependencies: depDetails,
		Message:      "Dependencies fetched successfully.",
	}, nil
}

func (m *ApplicationService) UpdateApplicationDependency(ctx context.Context, appUID string, input *model.UpdateApplicationDependencyRequest) (*model.UpdateApplicationDependencyResponse, error) {
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil || app == nil {
		return nil, fmt.Errorf("application not found")
	}

	var updated, failed []string
	for _, upd := range input.Updates {
		depID, err := uuid.Parse(upd.DependencyID)
		if err != nil {
			failed = append(failed, upd.DependencyID)
			continue
		}

		// Make sure at least UsedVersion is provided
		if upd.UsedVersion == "" {
			failed = append(failed, upd.DependencyID)
			continue
		}

		// Check if the app-dependency relationship exists
		appDep, err := m.appToDepedencyRepository.GetByAppAndDependencyID(ctx, appID, depID)
		if err != nil || appDep == nil {
			failed = append(failed, upd.DependencyID)
			continue
		}

		var versionCommitSHA string
		if upd.RepositoryURL != "" {
			// only fetch metadata if GitHub URL is provided
			parts, isValid := helper.ExtractGitHubOwnerRepo(upd.RepositoryURL)
			if isValid {
				// Fetch repo info to validate URL
				repoInfo, err := m.githubApiService.GetRepoInfo(parts.Owner, parts.Repo)
				if err == nil && repoInfo != nil {
					depedency, err := m.depedencyRepository.GetByID(ctx, appDep.DependencyID)
					if err == nil && depedency != nil {
						// Update repository URL if changed
						var version string
						versionCommitSHA, version, err = m.fetchAndUpdateDependencyMetadata(ctx, depedency, parts.Owner, parts.Repo, upd.UsedVersion, upd.RepositoryURL)
						if err == nil && version != "" {
							upd.UsedVersion = version // update to matched version if found
						}
					} else {
						slog.Warn("Dependency not found when updating metadata", "dependency_id", appDep.DependencyID)
					}
				} else {
					slog.Warn("Failed to fetch repository info from GitHub", "owner", parts.Owner, "repo", parts.Repo, "error", err)
				}
			} else {
				slog.Warn("Invalid GitHub URL provided, skipping metadata fetch", "url", upd.RepositoryURL)
			}
		}

		appDep.UsedVersion = upd.UsedVersion
		if versionCommitSHA != "" {
			appDep.UsedCommitSHA = &versionCommitSHA
		}
		if err := m.appToDepedencyRepository.Update(ctx, appDep); err != nil {
			failed = append(failed, upd.DependencyID)
			continue
		}
		updated = append(updated, upd.DependencyID)
	}

	msg := fmt.Sprintf("Updated: %d, Failed: %d", len(updated), len(failed))
	return &model.UpdateApplicationDependencyResponse{
		AppID:   appID.String(),
		Updated: updated,
		Failed:  failed,
		Message: msg,
	}, nil
}

func (m *ApplicationService) RemoveApplicationDependency(ctx context.Context, appUID string, deps []string) (interface{}, error) {
	// Parse app UUID
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	// Check if app exists
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	// Prepare
	results := make(map[string]interface{})
	successful := 0
	failed := 0

	// Process each dependency ID
	for _, depIDStr := range deps {
		depID, err := uuid.Parse(depIDStr)
		if err != nil {
			results[depIDStr] = map[string]interface{}{
				"status": "failed",
				"error":  "invalid dependency ID format",
			}
			failed++
			continue
		}

		// Check if the app-dependency relationship exists
		appDep, err := m.appToDepedencyRepository.GetByAppAndDependencyID(ctx, appID, depID)
		if err != nil && err != gorm.ErrRecordNotFound {
			results[depIDStr] = map[string]interface{}{
				"status": "failed",
				"error":  fmt.Sprintf("database error checking app dependency: %v", err),
			}
			failed++
			continue
		}
		if appDep == nil {
			results[depIDStr] = map[string]interface{}{
				"status": "skipped",
				"reason": "dependency not associated with this application",
			}
			continue
		}

		// Delete the app-dependency relationship
		err = m.appToDepedencyRepository.Delete(ctx, appDep.ID)
		if err != nil {
			results[depIDStr] = map[string]interface{}{
				"status": "failed",
				"error":  fmt.Sprintf("failed to remove app dependency: %v", err),
			}
			failed++
			continue
		}

		results[depIDStr] = map[string]interface{}{
			"status": "success",
		}
		successful++
	}

	return map[string]interface{}{
		"app_id":   app.ID.String(),
		"app_name": app.Name,
		"results":  results,
		"summary": map[string]interface{}{
			"total":      len(deps),
			"successful": successful,
			"failed":     failed,
		},
	}, nil
}

func (m *ApplicationService) RemoveApplication(ctx context.Context, appUID string) error {
	// Find the app by ID (UUID)
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return fmt.Errorf("invalid app ID: %w", err)
	}
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to fetch app: %w", err)
	}
	if app == nil {
		return fmt.Errorf("application not found")
	}

	return m.appRepository.UpdateStatus(ctx, appID, "inactive")
}

func (m *ApplicationService) RecoverApplication(ctx context.Context, appUID string) error {
	// Find the app by ID (UUID)
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return fmt.Errorf("invalid app ID: %w", err)
	}
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to fetch app: %w", err)
	}
	if app == nil {
		return fmt.Errorf("application not found")
	}
	return m.appRepository.UpdateStatus(ctx, appID, "active")
}

func (m *ApplicationService) ListApplications(ctx context.Context) (*model.ListApplicationsResponse, error) {
	apps, err := m.appRepository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}

	var summaries []model.ApplicationSummary
	for _, app := range apps {
		runtimeName := ""
		frameworkName := ""
		if app.RuntimeID != nil {
			runtime, _ := m.runTimeRepository.GetByID(ctx, *app.RuntimeID)
			if runtime != nil {
				runtimeName = runtime.Name
			}
		}
		if app.FrameworkID != nil {
			framework, _ := m.frameWorkRepository.GetByID(ctx, *app.FrameworkID)
			if framework != nil {
				frameworkName = framework.Name
			}
		}
		summaries = append(summaries, model.ApplicationSummary{
			AppID:       app.ID.String(),
			AppName:     app.Name,
			RuntimeType: runtimeName,
			Framework:   frameworkName,
			Status:      app.Status,
			Description: derefString(app.Description),
		})
	}

	return &model.ListApplicationsResponse{
		Applications: summaries,
		Message:      "Applications fetched successfully.",
	}, nil
}

func (m *ApplicationService) GetApplicationStatus(ctx context.Context, appUID string) (map[string]interface{}, error) {
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil || app == nil {
		return nil, fmt.Errorf("application not found")
	}
	appDeps, err := m.appToDepedencyRepository.GetByAppID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app dependencies: %w", err)
	}
	lastUpdated := ""
	if !app.UpdatedAt.IsZero() {
		lastUpdated = app.UpdatedAt.Format(time.RFC3339)
	}
	status := model.ApplicationStatus{
		AppID:           app.ID.String(),
		AppName:         app.Name,
		Status:          app.Status,
		DependencyCount: len(appDeps),
		LastUpdated:     lastUpdated,
	}
	return map[string]interface{}{"status": status}, nil
}

func (m *ApplicationService) ScanApplicationDependencies(ctx context.Context, appUID string) (interface{}, error) {
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}
	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil || app == nil {
		return nil, fmt.Errorf("application not found")
	}
	appDeps, err := m.appToDepedencyRepository.GetByAppID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app dependencies: %w", err)
	}

	runtime, err := m.runTimeRepository.GetByID(ctx, *app.RuntimeID)
	if err != nil || runtime == nil {
		return nil, fmt.Errorf("failed to fetch runtime info for application")
	}

	framework, err := m.frameWorkRepository.GetByID(ctx, *app.FrameworkID)
	frameworkName := ""
	if err == nil && framework != nil {
		frameworkName = framework.Name
	}

	var (
		wg            sync.WaitGroup
		mu            sync.Mutex
		findings      []model.ScanFinding
		depsWithVulns []helper.DependencyWithVulnerabilities
		totalCritical int
		totalHigh     int
		totalMedium   int
		totalLow      int
	)

	for _, appDep := range appDeps {
		wg.Add(1)
		go func(ad *entity.AppDependency) {
			defer wg.Done()
			dep, err := m.depedencyRepository.GetByID(ctx, ad.DependencyID)
			if err != nil || dep == nil || dep.Owner == "" || dep.Repo == "" {
				return
			}

			depInfo := parser.DependencyInfo{
				Name:         dep.Name,
				Owner:        dep.Owner,
				Repo:         dep.Repo,
				GitHubURL:    derefString(dep.RepositoryURL),
				Version:      ad.UsedVersion,
				IsGitHubRepo: dep.Owner != "" && dep.Repo != "",
				Runtime:      runtime.Name,
			}

			result, err := m.cveService.CheckDependencyVulnerabilities(ctx, depInfo)
			if err != nil {
				slog.Warn("Failed to check vulnerabilities", "dependency", dep.Name, "error", err)
				return
			}

			severity := "low" // default
			if result.CriticalCount > 0 {
				severity = "critical"
			} else if result.HighCount > 0 {
				severity = "high"
			} else if result.MediumCount > 0 {
				severity = "medium"
			} else if result.LowCount > 0 {
				severity = "low"
			}

			var vulnIDs []string
			for _, v := range result.Vulnerabilities {
				vulnIDs = append(vulnIDs, v.ID)
			}

			recommendation := ""
			if len(result.Recommendations) > 0 {
				recommendation = result.Recommendations[0]
			}

			finding := model.ScanFinding{
				Dependency:       dep.Name + ":" + dep.Repo,
				Version:          ad.UsedVersion,
				Severity:         severity,
				VulnerabilityIDs: vulnIDs,
				Recommendation:   recommendation,
			}

			// Create enhanced dependency with vulnerabilities for SBOM
			depWithVuln := helper.DependencyWithVulnerabilities{
				Name:            dep.Name,
				Version:         ad.UsedVersion,
				Owner:           dep.Owner,
				Repo:            dep.Repo,
				RepositoryURL:   derefString(dep.RepositoryURL),
				Runtime:         runtime.Name,
				IsGitHub:        dep.Owner != "" && dep.Repo != "",
				Vulnerabilities: result.Vulnerabilities,
				RiskScore:       result.RiskScore,
			}

			mu.Lock()
			findings = append(findings, finding)
			depsWithVulns = append(depsWithVulns, depWithVuln)
			totalCritical += result.CriticalCount
			totalHigh += result.HighCount
			totalMedium += result.MediumCount
			totalLow += result.LowCount
			mu.Unlock()
		}(appDep)
	}
	wg.Wait()

	summary := helper.AggregateVulnerabilitySummary(findings)
	failOn := []string{"high", "critical"}
	policyStatus, policyReason := helper.EvaluatePolicy(summary, failOn)

	artifacts := model.ScanArtifacts{
		VulnerabilityReport: fmt.Sprintf("https://your-app/api/scans/%s/report", app.ID.String()),
		SBOM:                fmt.Sprintf("https://your-app/api/scans/%s/sbom", app.ID.String()),
	}

	result := model.ScanApplicationResult{
		AppID:      app.ID.String(),
		AppName:    app.Name,
		ScanStatus: "completed",
		Summary:    summary,
		Policies:   model.ScanPolicy{FailOn: failOn, Status: policyStatus, Reason: policyReason},
		Artifacts:  artifacts,
		Findings:   findings,
	}

	// Generate enhanced SBOM from comprehensive vulnerability data
	enhancedSBOMData := helper.EnhancedSBOMData{
		AppID:   app.ID.String(),
		AppName: app.Name,
		// AppVersion:    "1.0.0", // You can fetch this from app metadata if available
		Runtime:       runtime.Name,
		Framework:     frameworkName,
		Dependencies:  depsWithVulns,
		ScanTimestamp: time.Now().UTC(),
		TotalFindings: len(findings),
		CriticalCount: totalCritical,
		HighCount:     totalHigh,
		MediumCount:   totalMedium,
		LowCount:      totalLow,
	}

	sbomBytes, err := helper.GenerateEnhancedCycloneDXSBOM(enhancedSBOMData)
	if err != nil {
		slog.Warn("Failed to generate enhanced SBOM", "error", err)
	} else {
		slog.Info("Enhanced SBOM generated successfully",
			"app_id", app.ID.String(),
			"size_bytes", len(sbomBytes),
			"total_components", len(depsWithVulns),
			"total_vulnerabilities", len(findings))

		// Save SBOM to object storage if service is available
		if m.objectStorageService != nil {
			sbomKey, err := m.objectStorageService.SaveSBOM(ctx, app.ID.String(), app.Name, sbomBytes, "json")
			if err != nil {
				slog.Error("Failed to save SBOM to object storage", "error", err)
			} else {
				slog.Info("SBOM saved to object storage successfully", "key", sbomKey)
				// Update the SBOM artifact URL with the actual storage key
				artifacts.SBOM = fmt.Sprintf("https://your-app/api/sbom/%s", sbomKey)
			}
		} else {
			slog.Warn("Object storage service not available, SBOM not persisted")
		}
	}

	return result, nil
}

func (m *ApplicationService) GetApplicationSBOM(ctx context.Context, appUID string) ([]byte, error) {
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil || app == nil {
		return nil, fmt.Errorf("application not found")
	}

	if m.objectStorageService == nil {
		return nil, fmt.Errorf("object storage service not available")
	}

	// List all SBOMs for this app
	sbomKeys, err := m.objectStorageService.ListSBOMs(ctx, app.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list SBOMs: %w", err)
	}

	if len(sbomKeys) == 0 {
		return nil, fmt.Errorf("no SBOM found for application")
	}

	// Get the latest SBOM (last in the list, assuming chronological order)
	latestSBOMKey := sbomKeys[len(sbomKeys)-1]
	sbomData, err := m.objectStorageService.GetSBOM(ctx, latestSBOMKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve SBOM: %w", err)
	}

	return sbomData, nil
}

func (m *ApplicationService) ListApplicationSBOMs(ctx context.Context, appUID string) ([]string, error) {
	appID, err := uuid.Parse(appUID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	app, err := m.appRepository.GetByID(ctx, appID)
	if err != nil || app == nil {
		return nil, fmt.Errorf("application not found")
	}

	if m.objectStorageService == nil {
		return nil, fmt.Errorf("object storage service not available")
	}

	sbomKeys, err := m.objectStorageService.ListSBOMs(ctx, app.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list SBOMs: %w", err)
	}

	return sbomKeys, nil
}

// processDependency processes a single dependency for an application
func (m *ApplicationService) processDependency(ctx context.Context, dep helper.DependencyInfo, app *entity.App, errCh chan<- error) {
	lookupOwner := dep.Owner
	lookupRepo := dep.Repo
	if lookupOwner == "" || lookupRepo == "" {
		lookupRepo = dep.Name // fallback for legacy/ambiguous cases
	}

	// Check if dependency already exists (case-insensitive)
	var dependency *entity.Dependency
	existingDep, err := m.depedencyRepository.GetByOwnerRepoCI(ctx, dep.Owner, dep.Repo)
	if err != nil && err != gorm.ErrRecordNotFound {
		errCh <- fmt.Errorf("failed to check existing dependency %s/%s: %w", dep.Owner, dep.Repo, err)
		return
	}

	// If not found, create new dependency
	var versionCommitSHA string

	if existingDep != nil {
		dependency = existingDep
	} else {
		// Create new dependency
		dependency = &entity.Dependency{
			ID:            uuid.New(),
			Name:          dep.Name,
			Owner:         dep.Owner,
			Repo:          dep.Repo,
			RepositoryURL: &dep.GitHubURL,
		}
		// err = m.depedencyRepository.Create(ctx, dependency)
		if err := m.depedencyRepository.Create(ctx, dependency); err != nil {
			// If unique constraint error, re-query and use existing
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "UNIQUE") {
				dependency, err = m.depedencyRepository.GetByOwnerRepoCI(ctx, lookupOwner, lookupRepo)
				if err != nil || dependency == nil {
					errCh <- fmt.Errorf("dependency create race: %w", err)
					return
				}
			} else {
				slog.Error("failed to create dependency", "error", err)
				errCh <- err
				return
			}
		}
		// If GitHub URL is valid, fetch and update metadata
		if dep.GitHubURL != "" {
			parts, isValid := helper.ExtractGitHubOwnerRepo(dep.GitHubURL)
			if isValid {
				var version string
				versionCommitSHA, version, err = m.fetchAndUpdateDependencyMetadata(ctx, dependency, parts.Owner, parts.Repo, dep.Version, dep.GitHubURL)
				if err == nil && version != "" {
					dep.Version = version // update to matched version if found
				}
			}
		}
	}

	// Check if app-dependency relationship already exists
	existingAppDep, err := m.appToDepedencyRepository.GetByAppAndDependencyID(ctx, app.ID, dependency.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		errCh <- fmt.Errorf("failed to check app dependency relationship: %w", err)
		return
	}

	if existingAppDep != nil {
		// Update version if different
		if existingAppDep.UsedVersion != dep.Version {
			existingAppDep.UsedVersion = dep.Version
			err = m.appToDepedencyRepository.Update(ctx, existingAppDep)
			if err != nil {
				errCh <- fmt.Errorf("failed to update app dependency version: %w", err)
				return
			}
		}
		return
	}

	// Create app-dependency relationship
	appDependency := &entity.AppDependency{
		ID:            uuid.New(),
		AppID:         app.ID,
		DependencyID:  dependency.ID,
		UsedVersion:   dep.Version,
		IsMonitored:   false,
		MonitorStatus: nil,
	}
	// Set UsedCommitSHA if we resolved it
	if versionCommitSHA != "" {
		appDependency.UsedCommitSHA = &versionCommitSHA
	}

	err = m.appToDepedencyRepository.Create(ctx, appDependency)
	if err != nil {
		errCh <- fmt.Errorf("failed to create app dependency: %w", err)
		return
	}

}

// fetchAndUpdateDependencyMetadata fetches GitHub metadata and updates the Dependency entity. Returns version commit SHA if found.
func (m *ApplicationService) fetchAndUpdateDependencyMetadata(ctx context.Context, dep *entity.Dependency, owner, repo, version, newRepoURL string) (string, string, error) {
	var defaultBranch, lastCommitSHA, lastCommitTime, latestTag string

	// Fetch default branch
	defaultBranch, err := m.githubApiService.GetDefaultBranch(owner, repo)
	if err != nil {
		slog.Error("failed to fetch default branch from GitHub", "error", err)
	}

	// Fetch latest commit
	listCommits, err := m.githubApiService.GetListCommits(owner, repo, defaultBranch)
	if err != nil {
		slog.Error("failed to fetch commits from GitHub", "error", err)
	}
	if len(listCommits) > 0 {
		commit := listCommits[0]
		lastCommitSHA, _ = commit["oid"].(string)
		lastCommitTime, _ = commit["author_date"].(string)
	}

	// Fetch tags
	listTags, err := m.githubApiService.ListTags(owner, repo)
	if err != nil {
		slog.Error("failed to fetch tags from GitHub", "error", err)
	}
	if len(listTags) > 0 {
		latestTag, _ = listTags[0]["name"].(string)
	}

	// find exact matching tag for the specified version
	matchingTag, err := m.githubApiService.FindMatchingTag(owner, repo, version)
	if err == nil && matchingTag != "" {
		version = matchingTag
	}

	// Get commit SHA for the specified version (tag/branch)
	var versionCommitSHA string
	shaCommit, isFound := helper.GetCommitSHAFromVersion(version, listTags)
	if isFound {
		versionCommitSHA = shaCommit
	}

	// Update Dependency entity fields
	if newRepoURL != "" {
		dep.RepositoryURL = &newRepoURL
	}
	dep.DefaultBranch = &defaultBranch
	dep.LastCommitSHA = &lastCommitSHA
	if lastCommitTime != "" {
		t, err := time.Parse(time.RFC3339, strings.ReplaceAll(lastCommitTime, " ", "T"))
		if err == nil {
			dep.LastCommitAt = &t
		}
	}
	dep.LastTag = &latestTag
	if err := m.depedencyRepository.Update(ctx, dep); err != nil {
		return versionCommitSHA, version, err
	}

	// Optionally, create a new DependencyVersion record
	if lastCommitSHA != "" {
		commitTime, _ := time.Parse(time.RFC3339, strings.ReplaceAll(lastCommitTime, " ", "T"))
		depVersion := &entity.DependencyVersion{
			ID:           uuid.New(),
			DependencyID: dep.ID,
			CommitSHA:    lastCommitSHA,
			CommitAt:     commitTime,
			Tag:          &latestTag,
			Branch:       &defaultBranch,
		}
		if err := m.depedencyVersionRepository.Create(ctx, depVersion); err != nil {
			slog.Error("failed to create dependency version", "error", err)
		}
	}

	return versionCommitSHA, version, nil
}

// auditApplicationAction audits application-related actions
func (m *ApplicationService) auditApplicationAction(ctx context.Context, appID uuid.UUID, action string, oldValues, newValues interface{}) error {
	return m.createAuditTrailEntry(ctx, "app", appID, action, oldValues, newValues, "user", false, nil)
}

// createAuditTrailEntry creates an audit trail entry for tracking monitoring activities
func (m *ApplicationService) createAuditTrailEntry(ctx context.Context, entityType string, entityID uuid.UUID, action string, oldValues, newValues interface{}, performedBy string, securityRelevant bool, riskLevel *string) error {
	if m.auditTrailRepository == nil {
		slog.Warn("Audit trail repository not available, skipping audit entry")
		return nil
	}

	// Marshal oldValues and newValues to JSON bytes
	var oldValuesBytes, newValuesBytes []byte
	var err error

	if oldValues != nil {
		oldValuesBytes, err = json.Marshal(oldValues)
		if err != nil {
			slog.Warn("Failed to marshal old values for audit trail", "error", err)
			oldValuesBytes = nil
		}
	}

	if newValues != nil {
		newValuesBytes, err = json.Marshal(newValues)
		if err != nil {
			slog.Warn("Failed to marshal new values for audit trail", "error", err)
			newValuesBytes = nil
		}
	}

	// Marshal context to JSON bytes
	contextData := map[string]interface{}{
		"service":    "monitoring_service_v2",
		"timestamp":  time.Now().UTC(),
		"session_id": uuid.New().String(), // Could be extracted from context
	}
	contextBytes, err := json.Marshal(contextData)
	if err != nil {
		slog.Warn("Failed to marshal context for audit trail", "error", err)
		contextBytes = nil
	}

	auditEntry := &entity.AuditTrail{
		ID:               uuid.New(),
		EntityType:       entityType,
		EntityID:         entityID,
		Action:           action,
		OldValues:        oldValuesBytes,
		NewValues:        newValuesBytes,
		PerformedBy:      performedBy,
		PerformedAt:      time.Now().UTC(),
		SecurityRelevant: securityRelevant,
		RiskLevel:        riskLevel,
		Context:          contextBytes,
	}

	return m.auditTrailRepository.Create(ctx, auditEntry)
}

// derefString safely dereferences a *string, returns "" if nil
func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// // createSecurityAuditEntry creates a security-focused audit trail entry
// func (m *ApplicationService) createSecurityAuditEntry(ctx context.Context, entityType string, entityID uuid.UUID, action string, details map[string]interface{}, riskLevel string) error {
// 	return m.createAuditTrailEntry(ctx, entityType, entityID, action, nil, details, "system", true, &riskLevel)
// }
