package services

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/helper"
	"elang-backend/internal/helper/parser"
	"elang-backend/internal/model"
	"elang-backend/internal/model/dto"
	"elang-backend/internal/repository"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MonitoringJobContext holds context for active monitoring jobs
type MonitoringJobContext struct {
	Job        *entity.MonitoringJob
	CancelFunc context.CancelFunc
	StopChan   chan struct{}
	Progress   *JobProgress
	// mutex      sync.RWMutex
}

// JobProgress tracks real-time progress of monitoring jobs
type JobProgress struct {
	TotalChecks        int            `json:"total_checks"`
	CompletedChecks    int            `json:"completed_checks"`
	FailedChecks       int            `json:"failed_checks"`
	SecurityDetections int            `json:"security_detections"`
	StartTime          time.Time      `json:"start_time"`
	LastUpdate         time.Time      `json:"last_update"`
	CurrentOperation   string         `json:"current_operation"`
	EstimatedTimeLeft  *time.Duration `json:"estimated_time_left"`
}

type DependenciesService struct {
	depedencyParserService helper.DependencyParser
	cveService             *helper.CVEHelper
	objectStorageService   ObjectStorageInterface
	sharedScanner          *helper.SharedScanner

	appRepository       repository.ApplicationRepository
	depedencyRepository repository.DependencyRepository
	appDepedencyRepo    repository.AppDependencyRepository
	runTimeRepository   repository.RuntimeRepository

	activeJobs   map[uuid.UUID]*MonitoringJobContext // Save active monitoring jobs
	jobsMutex    sync.RWMutex                        // Mutex to protect access to activeJobs
	shutdownChan chan struct{}                       // Channel to signal shutdown
	workerPool   chan struct{}                       // For controlling concurrency
}

func NewDependenciesService(basicRepo dto.BasicRepositories, objectStorageService ObjectStorageInterface, dependencyParser helper.DependencyParser) DependenciesInterface {
	return &DependenciesService{
		depedencyParserService: dependencyParser,
		cveService:             helper.NewCVEHelper(),
		sharedScanner:          helper.NewSharedScanner(10), // default max 10 concurrent scans
		activeJobs:             make(map[uuid.UUID]*MonitoringJobContext),
		shutdownChan:           make(chan struct{}),
		workerPool:             make(chan struct{}, 5), // default max 5 concurrent jobs

		objectStorageService: objectStorageService,

		appRepository:       basicRepo.AppRepository,
		depedencyRepository: basicRepo.DepedencyRepository,
		appDepedencyRepo:    basicRepo.AppToDepedencyRepository,
		runTimeRepository:   basicRepo.RunTimeRepository,
	}
}

func (s *DependenciesService) ScanDependencies(ctx context.Context, appName, runtime, version, description, fileName, content string) (interface{}, error) {
	// Implementation for scanning application dependencies
	if appName == "" || content == "" || runtime == "" {
		return nil, fmt.Errorf("appName, version, and content are required")
	}

	if !isRuntimeSupported(runtime) {
		return nil, fmt.Errorf("runtime %s is not supported", runtime)
	}

	// Parse dependencies from the provided content
	deps := s.depedencyParserService.ParseDependencyFile(fileName, content, parser.RuntimeType(runtime))
	if len(deps.Dependencies) == 0 {
		return nil, fmt.Errorf("no dependencies found in the provided content")
	}

	findings, depsWithVulns, totalCritical, totalHigh, totalMedium, totalLow := s.sharedScanner.ScanDependenciesWithControl(ctx, deps.Dependencies)

	// START SCANNING PROCESS
	// TEMPORARY: Using previous scanning logic for reference
	// END SCANNING PROCESS

	// Aggregate summary and evaluate policies
	summary := helper.AggregateVulnerabilitySummary(findings)
	failOn := []string{"high", "critical"}
	policyStatus, policyReason := helper.EvaluatePolicy(summary, failOn)

	scanID := uuid.New().String()

	artifacts := model.ScanArtifacts{
		VulnerabilityReport: fmt.Sprintf("https://your-app/api/scans/%s/report", scanID),
		SBOM:                fmt.Sprintf("https://your-app/api/scans/%s/sbom", scanID),
	}

	result := model.ScanApplicationResult{
		AppID:      scanID,
		AppName:    appName,
		ScanStatus: "completed",
		Summary:    summary,
		Policies:   model.ScanPolicy{FailOn: failOn, Status: policyStatus, Reason: policyReason},
		Artifacts:  artifacts,
		Findings:   findings,
	}

	// Generate enhanced SBOM from comprehensive vulnerability data
	enhancedSBOMData := helper.EnhancedSBOMData{
		AppID:         scanID,
		AppName:       appName,
		AppVersion:    version, // You can fetch this from app metadata if available
		Runtime:       runtime,
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
			"app_id", scanID,
			"size_bytes", len(sbomBytes),
			"total_components", len(depsWithVulns),
			"total_vulnerabilities", len(findings))

		// Save SBOM to object storage if service is available
		if s.objectStorageService != nil {
			sbomKey, err := s.objectStorageService.SaveSBOM(ctx, scanID, appName, sbomBytes, "json")
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

func (s *DependenciesService) GetSBOMById(ctx context.Context, appName, scanID string) ([]byte, error) {
	// Input validation
	if scanID == "" || appName == "" {
		return nil, fmt.Errorf("appName and scanID are required")
	}
	if s.objectStorageService == nil {
		return nil, fmt.Errorf("object storage service not available")
	}

	// List all SBOMs for the app
	sbomKeys, err := s.objectStorageService.ListSBOMs(ctx, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to list SBOMs: %w", err)
	}

	// Find the SBOM key that contains the scanID
	var targetKey string
	for _, key := range sbomKeys {
		if strings.Contains(key, scanID) {
			targetKey = key
			break
		}
	}
	if targetKey == "" {
		return nil, fmt.Errorf("SBOM not found for scanID: %s", scanID)
	}

	// Retrieve the SBOM
	sbomData, err := s.objectStorageService.GetSBOM(ctx, targetKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve SBOM: %w", err)
	}
	return sbomData, nil
}

func (s *DependenciesService) StartMonitoringApplication(ctx context.Context, appID string) error {
	// Implementation for starting monitoring an application
	app, err := s.getAppByID(ctx, appID)
	if err != nil {
		return err
	}
	runtime, err := s.runTimeRepository.GetByID(ctx, *app.RuntimeID)
	if err != nil {
		return fmt.Errorf("failed to get runtime: %w", err)
	}

	// Here you would add logic to start monitoring the application,
	go func() {
		jobID := uuid.New()
		stopChan := make(chan struct{}) // Create a stop channel

		// Create and register the monitoring job context
		jobContext := &MonitoringJobContext{
			Job: &entity.MonitoringJob{
				ID:        jobID,
				AppIDs:    []uuid.UUID{app.ID},
				Status:    "running",
				CreatedAt: time.Now(),
				CreatedBy: "system",
			},
			Progress: &JobProgress{
				CompletedChecks:    0,
				FailedChecks:       0,
				SecurityDetections: 0,
				StartTime:          time.Now(),
				LastUpdate:         time.Now(),
				CurrentOperation:   "initializing",
			},
			StopChan: stopChan,
		}
		// Add to active jobs with logging
		s.jobsMutex.Lock()
		s.activeJobs[jobID] = jobContext
		s.jobsMutex.Unlock() // Unlock after adding
		slog.Info("Monitoring job started",
			"job_id", jobID.String(),
			"app_id", app.ID.String(),
			"active_jobs_count", len(s.activeJobs))

		// Monitoring loop with proper cleanup
		defer func() {
			s.jobsMutex.Lock()
			delete(s.activeJobs, jobID) // Remove job from active jobs if exists
			s.jobsMutex.Unlock()
			slog.Info("Monitoring job cleaned up", "job_id", jobID.String())
		}()

		// Periodic monitoring task
		ticker := time.NewTicker(24 * time.Hour)
		// ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-stopChan:
				slog.Info("Monitoring job stopped", "job_id", jobID.String(), "app_id", app.ID.String())
				return
			case <-ticker.C:
				slog.Info("Monitoring application dependencies", "app_id", appID, "app_name", app.Name)

				jobContext.Progress.CurrentOperation = "scanning"
				jobContext.Progress.LastUpdate = time.Now()
				jobContext.Progress.FailedChecks = 0

				context := context.Background()

				// Here you would implement the actual monitoring logic,
				appDeps, err := s.appDepedencyRepo.GetByAppID(context, app.ID)
				if err != nil {
					slog.Error("Failed to get app dependencies", "error", err)
					jobContext.Progress.FailedChecks++
					continue
				}
				if len(appDeps) == 0 {
					slog.Warn("No dependencies found for the application", "app_id", appID)
					continue
				}

				// Fetch dependency details
				var depedenciesInfoList []parser.DependencyInfo
				for _, dep := range appDeps {
					depedenciesData, err := s.depedencyRepository.GetByID(context, dep.DependencyID)
					if err != nil {
						slog.Error("Failed to get dependency", "error", err)
						jobContext.Progress.FailedChecks++
						continue
					}
					if depedenciesData == nil {
						slog.Error("Dependency not found", "dependency_id", dep.DependencyID.String())
						jobContext.Progress.FailedChecks++
						continue
					}
					depedenciesInfoList = append(depedenciesInfoList, parser.DependencyInfo{
						Name:    depedenciesData.Name,
						Version: dep.UsedVersion,
						Runtime: runtime.Name,
					})
				}

				// Perform scanning with controlled concurrency
				findings, depsWithVulns, totalCritical, totalHigh, totalMedium, totalLow := s.sharedScanner.ScanDependenciesWithControl(context, depedenciesInfoList)
				jobContext.Progress.CompletedChecks = len(findings)

				// Aggregate summary and evaluate policies
				summary := helper.AggregateVulnerabilitySummary(findings)
				failOn := []string{"high", "critical"}
				policyStatus, policyReason := helper.EvaluatePolicy(summary, failOn)

				// Generate a unique scan ID for this monitoring scan
				scanID := uuid.New().String()
				artifacts := model.ScanArtifacts{
					VulnerabilityReport: fmt.Sprintf("https://your-app/api/scans/%s/report", scanID),
					SBOM:                fmt.Sprintf("https://your-app/api/scans/%s/sbom", scanID),
				}

				result := model.ScanApplicationResult{
					AppID:      scanID,
					AppName:    app.Name,
					ScanStatus: "completed",
					Summary:    summary,
					Policies:   model.ScanPolicy{FailOn: failOn, Status: policyStatus, Reason: policyReason},
					Artifacts:  artifacts,
					Findings:   findings,
				}
				_ = result // You can store or process the result as needed

				// Generate enhanced SBOM from comprehensive vulnerability data
				enhancedSBOMData := helper.EnhancedSBOMData{
					AppID:         scanID,
					AppName:       app.Name,
					Runtime:       runtime.Name,
					Dependencies:  depsWithVulns,
					ScanTimestamp: time.Now().UTC(),
					TotalFindings: len(findings),
					CriticalCount: totalCritical,
					HighCount:     totalHigh,
					MediumCount:   totalMedium,
					LowCount:      totalLow,
					// AppVersion:    , // You can fetch this from app metadata if available
				}
				sbomBytes, err := helper.GenerateEnhancedCycloneDXSBOM(enhancedSBOMData)
				if err != nil {
					slog.Error("Failed to generate enhanced SBOM", "error", err)
				} else {
					slog.Info("Enhanced SBOM generated successfully",
						"app_id", scanID,
						"size_bytes", len(sbomBytes),
						"total_components", len(depsWithVulns),
						"total_vulnerabilities", len(findings))
				}
				if s.objectStorageService != nil {
					sbomKey, err := s.objectStorageService.SaveSBOM(context, scanID, app.Name, sbomBytes, "json")
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
				slog.Info("Monitoring scan completed",
					"app_id", appID,
					"app_name", app.Name,
					"scan_id", scanID,
					"findings", len(findings),
					"critical", totalCritical,
					"high", totalHigh,
					"medium", totalMedium,
					"low", totalLow,
				)
				jobContext.Progress.LastUpdate = time.Now()
				jobContext.Progress.CurrentOperation = "idle"
			}
		}
	}()
	// such as scheduling periodic scans or setting up webhooks.
	slog.Info("Started monitoring application", "app_id", appID, "app_name", app.Name)
	return nil
}

func (s *DependenciesService) StopMonitoringApplication(ctx context.Context, appID string) error {
	app, err := s.getAppByID(ctx, appID)
	if err != nil {
		return err
	}
	// Find and stop the monitoring job
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()
	for jobID, jobCtx := range s.activeJobs {
		for _, id := range jobCtx.Job.AppIDs {
			if id == app.ID {
				close(jobCtx.StopChan) // Signal stop
				delete(s.activeJobs, jobID)
				slog.Info("Stopped monitoring application", "app_id", appID, "app_name", app.Name)
				return nil
			}
		}
	}
	return fmt.Errorf("monitoring job not found for app_id: %s", appID)
}

func (s *DependenciesService) GetMonitoringStatus(ctx context.Context, appID string) (map[string]interface{}, error) {
	app, err := s.getAppByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	// Check if app is being monitored
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	for _, jobCtx := range s.activeJobs {
		for _, id := range jobCtx.Job.AppIDs {
			if id == app.ID {
				status := map[string]interface{}{
					"app_uid":       appID,
					"monitoring":    true,
					"job_id":        jobCtx.Job.ID.String(),
					"status":        jobCtx.Job.Status,
					"started_at":    jobCtx.Job.CreatedAt,
					"last_checked":  jobCtx.Progress.LastUpdate,
					"next_check_in": "24 hours",
				}
				return status, nil
			}
		}
	}

	// Not being monitored
	status := map[string]interface{}{
		"app_uid":       appID,
		"monitoring":    false,
		"last_checked":  nil,
		"next_check_in": nil,
	}
	return status, nil
}

func isRuntimeSupported(runtime string) bool {
	runtime = strings.ToLower(runtime)
	supportedRuntimes := []string{"node.js", "python", "java", "go", "ruby", "php", "dotnet", "gradle"}
	for _, r := range supportedRuntimes {
		if r == runtime {
			return true
		}
	}
	return false
}

func (s *DependenciesService) getAppByID(ctx context.Context, appID string) (*entity.App, error) {
	// Implementation for starting monitoring an application
	appUID, err := uuid.Parse(appID)
	if err != nil {
		return nil, fmt.Errorf("invalid appID: %w", err)
	}

	app, err := s.appRepository.GetByID(ctx, appUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return nil, fmt.Errorf("application not found")
	}
	return app, nil
}
