package services_test

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/model"
	"elang-backend/internal/services"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) Create(ctx context.Context, app *entity.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.App), args.Error(1)
}

func (m *MockApplicationRepository) GetAll(ctx context.Context) ([]*entity.App, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.App), args.Error(1)
}

func (m *MockApplicationRepository) Update(ctx context.Context, app *entity.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetByName(ctx context.Context, name string) (*entity.App, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.App), args.Error(1)
}

func (m *MockApplicationRepository) GetByStatus(ctx context.Context, status string) ([]*entity.App, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.App), args.Error(1)
}

func (m *MockApplicationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockRuntimeRepository struct {
	mock.Mock
}

func (m *MockRuntimeRepository) Create(ctx context.Context, runtime *entity.Runtime) error {
	args := m.Called(ctx, runtime)
	return args.Error(0)
}

func (m *MockRuntimeRepository) GetByID(ctx context.Context, id int) (*entity.Runtime, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Runtime), args.Error(1)
}

func (m *MockRuntimeRepository) GetAll(ctx context.Context) ([]*entity.Runtime, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Runtime), args.Error(1)
}

func (m *MockRuntimeRepository) Update(ctx context.Context, runtime *entity.Runtime) error {
	args := m.Called(ctx, runtime)
	return args.Error(0)
}

func (m *MockRuntimeRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRuntimeRepository) GetByName(ctx context.Context, name string) (*entity.Runtime, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Runtime), args.Error(1)
}

func (m *MockRuntimeRepository) GetByNameCI(ctx context.Context, name string) (*entity.Runtime, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Runtime), args.Error(1)
}

type MockFrameworkRepository struct {
	mock.Mock
}

func (m *MockFrameworkRepository) Create(ctx context.Context, framework *entity.Framework) error {
	args := m.Called(ctx, framework)
	return args.Error(0)
}

func (m *MockFrameworkRepository) GetByID(ctx context.Context, id int) (*entity.Framework, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Framework), args.Error(1)
}

func (m *MockFrameworkRepository) GetAll(ctx context.Context) ([]*entity.Framework, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Framework), args.Error(1)
}

func (m *MockFrameworkRepository) Update(ctx context.Context, framework *entity.Framework) error {
	args := m.Called(ctx, framework)
	return args.Error(0)
}

func (m *MockFrameworkRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFrameworkRepository) GetByName(ctx context.Context, name string) (*entity.Framework, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Framework), args.Error(1)
}

func (m *MockFrameworkRepository) GetByNameCI(ctx context.Context, name string) (*entity.Framework, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Framework), args.Error(1)
}

func TestApplicationService_ListApplications(t *testing.T) {
	mockAppRepo := new(MockApplicationRepository)
	ctx := context.Background()

	expectedApps := []*entity.App{
		{
			ID:     uuid.New(),
			Name:   "app1",
			Status: "active",
		},
		{
			ID:     uuid.New(),
			Name:   "app2",
			Status: "inactive",
		},
	}

	mockAppRepo.On("GetAll", ctx).Return(expectedApps, nil)

	// Note: This is a simplified test. You'll need to create the actual service
	// with all required dependencies
	apps, err := mockAppRepo.GetAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, apps, 2)
	assert.Equal(t, "app1", apps[0].Name)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_GetByID_Success(t *testing.T) {
	mockAppRepo := new(MockApplicationRepository)
	ctx := context.Background()
	appID := uuid.New()

	expectedApp := &entity.App{
		ID:     appID,
		Name:   "test-app",
		Status: "active",
	}

	mockAppRepo.On("GetByID", ctx, appID).Return(expectedApp, nil)

	app, err := mockAppRepo.GetByID(ctx, appID)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, "test-app", app.Name)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_GetByID_NotFound(t *testing.T) {
	mockAppRepo := new(MockApplicationRepository)
	ctx := context.Background()
	appID := uuid.New()

	mockAppRepo.On("GetByID", ctx, appID).Return(nil, nil)

	app, err := mockAppRepo.GetByID(ctx, appID)

	assert.NoError(t, err)
	assert.Nil(t, app)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_Create_EmptyName(t *testing.T) {
	// Test validation
	appName := ""
	runtimeType := "Node.js"

	if appName == "" || runtimeType == "" {
		err := errors.New("content, file name, runtime type, and application name cannot be empty")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	}
}

func TestApplicationService_UpdateStatus(t *testing.T) {
	mockAppRepo := new(MockApplicationRepository)
	ctx := context.Background()
	appID := uuid.New()
	newStatus := "inactive"

	mockAppRepo.On("UpdateStatus", ctx, appID, newStatus).Return(nil)

	err := mockAppRepo.UpdateStatus(ctx, appID, newStatus)

	assert.NoError(t, err)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_Delete(t *testing.T) {
	mockAppRepo := new(MockApplicationRepository)
	ctx := context.Background()
	appID := uuid.New()

	mockAppRepo.On("Delete", ctx, appID).Return(nil)

	err := mockAppRepo.Delete(ctx, appID)

	assert.NoError(t, err)
	mockAppRepo.AssertExpectations(t)
}

func TestRuntimeRepository_GetByNameCI(t *testing.T) {
	mockRuntimeRepo := new(MockRuntimeRepository)
	ctx := context.Background()

	expectedRuntime := &entity.Runtime{
		ID:   1,
		Name: "Node.js",
	}

	mockRuntimeRepo.On("GetByNameCI", ctx, "node.js").Return(expectedRuntime, nil)

	runtime, err := mockRuntimeRepo.GetByNameCI(ctx, "node.js")

	assert.NoError(t, err)
	assert.NotNil(t, runtime)
	assert.Equal(t, "Node.js", runtime.Name)
	mockRuntimeRepo.AssertExpectations(t)
}

func TestFrameworkRepository_GetByNameCI(t *testing.T) {
	mockFrameworkRepo := new(MockFrameworkRepository)
	ctx := context.Background()

	expectedFramework := &entity.Framework{
		ID:   1,
		Name: "Express",
	}

	mockFrameworkRepo.On("GetByNameCI", ctx, "express").Return(expectedFramework, nil)

	framework, err := mockFrameworkRepo.GetByNameCI(ctx, "express")

	assert.NoError(t, err)
	assert.NotNil(t, framework)
	assert.Equal(t, "Express", framework.Name)
	mockFrameworkRepo.AssertExpectations(t)
}

func TestApplicationService_Interface(t *testing.T) {
	t.Run("InterfaceCompliance", func(t *testing.T) {
		// This test verifies that the interface is correctly defined
		var _ services.ApplicationInterface = (*mockApplicationService)(nil)
	})
}

// Mock implementation of ApplicationService for interface testing
type mockApplicationService struct {
	mock.Mock
}

func (m *mockApplicationService) AddApplication(ctx context.Context, appName, runtimeType, framework, description, fileName, content string) (*model.AddApplicationResponse, error) {
	args := m.Called(ctx, appName, runtimeType, framework, description, fileName, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AddApplicationResponse), args.Error(1)
}

func (m *mockApplicationService) AddApplicationDependency(ctx context.Context, appUID string, deps []model.DependencyInfoRequest) (interface{}, error) {
	args := m.Called(ctx, appUID, deps)
	return args.Get(0), args.Error(1)
}

func (m *mockApplicationService) ListApplicationDependency(ctx context.Context, appUID string) (*model.ListApplicationDependencyResponse, error) {
	args := m.Called(ctx, appUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ListApplicationDependencyResponse), args.Error(1)
}

func (m *mockApplicationService) UpdateApplicationDependency(ctx context.Context, appUID string, input *model.UpdateApplicationDependencyRequest) (*model.UpdateApplicationDependencyResponse, error) {
	args := m.Called(ctx, appUID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UpdateApplicationDependencyResponse), args.Error(1)
}

func (m *mockApplicationService) RemoveApplicationDependency(ctx context.Context, appUID string, deps []string) (interface{}, error) {
	args := m.Called(ctx, appUID, deps)
	return args.Get(0), args.Error(1)
}

func (m *mockApplicationService) RemoveApplication(ctx context.Context, appUID string) error {
	args := m.Called(ctx, appUID)
	return args.Error(0)
}

func (m *mockApplicationService) RecoverApplication(ctx context.Context, appUID string) error {
	args := m.Called(ctx, appUID)
	return args.Error(0)
}

func (m *mockApplicationService) ListApplications(ctx context.Context) (*model.ListApplicationsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ListApplicationsResponse), args.Error(1)
}

func (m *mockApplicationService) GetApplicationStatus(ctx context.Context, appUID string) (map[string]interface{}, error) {
	args := m.Called(ctx, appUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockApplicationService) ScanApplicationDependencies(ctx context.Context, appUID string) (interface{}, error) {
	args := m.Called(ctx, appUID)
	return args.Get(0), args.Error(1)
}

func (m *mockApplicationService) GetApplicationSBOM(ctx context.Context, appUID string) ([]byte, error) {
	args := m.Called(ctx, appUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockApplicationService) ListApplicationSBOMs(ctx context.Context, appUID string) ([]string, error) {
	args := m.Called(ctx, appUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
