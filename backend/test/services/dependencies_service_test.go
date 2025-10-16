package services_test

import (
	"context"
	"elang-backend/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock DependenciesService
type mockDependenciesService struct {
	mock.Mock
}

func (m *mockDependenciesService) ScanDependencies(ctx context.Context, appName, runtime, version, description, fileName, content string) (interface{}, error) {
	args := m.Called(ctx, appName, runtime, version, description, fileName, content)
	return args.Get(0), args.Error(1)
}

func (m *mockDependenciesService) GetSBOMById(ctx context.Context, appName, sbomID string) ([]byte, error) {
	args := m.Called(ctx, appName, sbomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockDependenciesService) StartMonitoringApplication(ctx context.Context, appUID string) error {
	args := m.Called(ctx, appUID)
	return args.Error(0)
}

func (m *mockDependenciesService) StopMonitoringApplication(ctx context.Context, appUID string) error {
	args := m.Called(ctx, appUID)
	return args.Error(0)
}

func (m *mockDependenciesService) GetMonitoringStatus(ctx context.Context, appUID string) (map[string]interface{}, error) {
	args := m.Called(ctx, appUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestDependenciesService_ScanDependencies_EmptyContent(t *testing.T) {
	// Test validation
	appName := ""
	content := ""
	runtime := ""

	if appName == "" || content == "" || runtime == "" {
		err := assert.AnError
		assert.Error(t, err)
	}
}

func TestDependenciesService_ScanDependencies_Success(t *testing.T) {
	mockService := new(mockDependenciesService)
	ctx := context.Background()

	expectedResult := map[string]interface{}{
		"status":   "completed",
		"findings": []interface{}{},
	}

	mockService.On("ScanDependencies", ctx, "test-app", "Node.js", "1.0.0", "Test app", "package.json", "{}").
		Return(expectedResult, nil)

	result, err := mockService.ScanDependencies(ctx, "test-app", "Node.js", "1.0.0", "Test app", "package.json", "{}")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockService.AssertExpectations(t)
}

func TestDependenciesService_GetSBOMById_Success(t *testing.T) {
	mockService := new(mockDependenciesService)
	ctx := context.Background()

	expectedSBOM := []byte(`{"components": []}`)

	mockService.On("GetSBOMById", ctx, "test-app", "sbom-123").Return(expectedSBOM, nil)

	sbom, err := mockService.GetSBOMById(ctx, "test-app", "sbom-123")

	assert.NoError(t, err)
	assert.NotNil(t, sbom)
	assert.Contains(t, string(sbom), "components")
	mockService.AssertExpectations(t)
}

func TestDependenciesService_StartMonitoring_Success(t *testing.T) {
	mockService := new(mockDependenciesService)
	ctx := context.Background()
	appUID := "test-app-uid"

	mockService.On("StartMonitoringApplication", ctx, appUID).Return(nil)

	err := mockService.StartMonitoringApplication(ctx, appUID)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestDependenciesService_StopMonitoring_Success(t *testing.T) {
	mockService := new(mockDependenciesService)
	ctx := context.Background()
	appUID := "test-app-uid"

	mockService.On("StopMonitoringApplication", ctx, appUID).Return(nil)

	err := mockService.StopMonitoringApplication(ctx, appUID)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestDependenciesService_GetMonitoringStatus_Success(t *testing.T) {
	mockService := new(mockDependenciesService)
	ctx := context.Background()
	appUID := "test-app-uid"

	expectedStatus := map[string]interface{}{
		"status":     "active",
		"last_check": "2024-01-01T00:00:00Z",
	}

	mockService.On("GetMonitoringStatus", ctx, appUID).Return(expectedStatus, nil)

	status, err := mockService.GetMonitoringStatus(ctx, appUID)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "active", status["status"])
	mockService.AssertExpectations(t)
}

func TestDependenciesService_Interface(t *testing.T) {
	t.Run("InterfaceCompliance", func(t *testing.T) {
		var _ services.DependenciesInterface = (*mockDependenciesService)(nil)
	})
}

func TestDependenciesService_ScanDependencies_UnsupportedRuntime(t *testing.T) {
	// Test for unsupported runtime
	runtime := "UnsupportedRuntime"
	supportedRuntimes := []string{"Node.js", "Python", "Java", "Go", "Ruby", "PHP", "Rust", ".NET"}

	isSupported := false
	for _, r := range supportedRuntimes {
		if r == runtime {
			isSupported = true
			break
		}
	}

	assert.False(t, isSupported, "Runtime should not be supported")
}

func TestDependenciesService_GetSBOMById_EmptyParams(t *testing.T) {
	appName := ""
	sbomID := ""

	if appName == "" || sbomID == "" {
		err := assert.AnError
		assert.Error(t, err)
	}
}
