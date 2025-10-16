package repository_test

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate all entities
	err = db.AutoMigrate(
		&entity.Runtime{},
		&entity.Framework{},
		&entity.App{},
		&entity.Dependency{},
		&entity.AppDependency{},
		&entity.DependencyVersion{},
		&entity.AuditTrail{},
	)
	require.NoError(t, err)

	return db
}

func TestApplicationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	app := &entity.App{
		ID:          uuid.New(),
		Name:        "test-app",
		Description: stringPtr("Test application"),
		Status:      "active",
	}

	err := repo.Create(ctx, app)
	assert.NoError(t, err)

	// Verify creation
	found, err := repo.GetByID(ctx, app.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, app.Name, found.Name)
}

func TestApplicationRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	app := &entity.App{
		ID:     uuid.New(),
		Name:   "test-app",
		Status: "active",
	}
	err := repo.Create(ctx, app)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, app.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, app.ID, found.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetByID(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestApplicationRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	apps := []*entity.App{
		{ID: uuid.New(), Name: "app1", Status: "active"},
		{ID: uuid.New(), Name: "app2", Status: "inactive"},
	}

	for _, app := range apps {
		err := repo.Create(ctx, app)
		require.NoError(t, err)
	}

	result, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestApplicationRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	app := &entity.App{
		ID:     uuid.New(),
		Name:   "test-app",
		Status: "active",
	}
	err := repo.Create(ctx, app)
	require.NoError(t, err)

	app.Name = "updated-app"
	err = repo.Update(ctx, app)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, app.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-app", found.Name)
}

func TestApplicationRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	app := &entity.App{
		ID:     uuid.New(),
		Name:   "test-app",
		Status: "active",
	}
	err := repo.Create(ctx, app)
	require.NoError(t, err)

	err = repo.Delete(ctx, app.ID)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, app.ID)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestApplicationRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	app := &entity.App{
		ID:     uuid.New(),
		Name:   "unique-app",
		Status: "active",
	}
	err := repo.Create(ctx, app)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByName(ctx, "unique-app")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, app.ID, found.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetByName(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestApplicationRepository_GetByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	apps := []*entity.App{
		{ID: uuid.New(), Name: "app1", Status: "active"},
		{ID: uuid.New(), Name: "app2", Status: "active"},
		{ID: uuid.New(), Name: "app3", Status: "inactive"},
	}

	for _, app := range apps {
		err := repo.Create(ctx, app)
		require.NoError(t, err)
	}

	activeApps, err := repo.GetByStatus(ctx, "active")
	assert.NoError(t, err)
	assert.Len(t, activeApps, 2)
}

func TestApplicationRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAppRepository(db)
	ctx := context.Background()

	app := &entity.App{
		ID:     uuid.New(),
		Name:   "test-app",
		Status: "active",
	}
	err := repo.Create(ctx, app)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, app.ID, "inactive")
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, app.ID)
	assert.NoError(t, err)
	assert.Equal(t, "inactive", found.Status)
}

func stringPtr(s string) *string {
	return &s
}
