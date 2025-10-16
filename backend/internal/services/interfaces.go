package services

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/model"
)

type ApplicationInterface interface {
	// Add or intialize Application -> input app name , depedency file , runtime type , description
	AddApplication(ctx context.Context, appName, runtimeType, framework, description, fileName, content string) (*model.AddApplicationResponse, error)

	// Add depedency to Application (batch)
	AddApplicationDependency(ctx context.Context, appUID string, deps []model.DependencyInfoRequest) (interface{}, error)

	// List Applications Dependency
	ListApplicationDependency(ctx context.Context, appUID string) (*model.ListApplicationDependencyResponse, error)

	// Update Application Dependency
	UpdateApplicationDependency(ctx context.Context, appUID string, input *model.UpdateApplicationDependencyRequest) (*model.UpdateApplicationDependencyResponse, error)

	// Remove depedency from Application (batch)
	RemoveApplicationDependency(ctx context.Context, appUID string, deps []string) (interface{}, error)

	// Remove Application or Deactivate Application
	RemoveApplication(ctx context.Context, appUID string) error

	// Recover Application or Reactivate Application
	RecoverApplication(ctx context.Context, appUID string) error

	// List Applications
	ListApplications(ctx context.Context) (*model.ListApplicationsResponse, error)

	// // Get Monitoring Status of Application
	GetApplicationStatus(ctx context.Context, appUID string) (map[string]interface{}, error)

	ScanApplicationDependencies(ctx context.Context, appUID string) (interface{}, error)

	// Get SBOM for an application
	GetApplicationSBOM(ctx context.Context, appUID string) ([]byte, error)

	// List all SBOMs for an application
	ListApplicationSBOMs(ctx context.Context, appUID string) ([]string, error)

	// // Get Monitoring Status of All Applications
	// GetAllApplicationsStatus(ctx context.Context) (map[string]interface{}, error)
}

type DependenciesInterface interface {
	// Scan Application for vulnerabilities by checking dependency versions in OSV
	ScanDependencies(ctx context.Context, appName, runtime, version, description, fileName, content string) (interface{}, error)

	// Get SBOM by its ID
	GetSBOMById(ctx context.Context, appName, sbomID string) ([]byte, error)

	// Start monitoring an application
	StartMonitoringApplication(ctx context.Context, appUID string) error

	// Stop monitoring an application
	StopMonitoringApplication(ctx context.Context, appUID string) error

	// Get monitoring status of an application
	GetMonitoringStatus(ctx context.Context, appUID string) (map[string]interface{}, error)
}

type DepedencyMonitoringInterface interface {
	// MonitorApplicationDepedencies starts monitoring an application's dependencies for changes
	MonitorApplicationDepedencies(ctx context.Context, app *entity.App) (interface{}, error)

	// StopMonitoringApplication stops monitoring an application's dependencies
	StopMonitoringApplication(ctx context.Context, app *entity.App) error

	// GetMonitoringStatus retrieves the monitoring status of an application
	GetMonitoringStatus(ctx context.Context, app *entity.App) (map[string]interface{}, error)
}
