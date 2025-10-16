package repository_test

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	dep := &entity.Dependency{
		ID:    uuid.New(),
		Name:  "test-dependency",
		Owner: "testowner",
		Repo:  "testrepo",
	}

	err := repo.Create(ctx, dep)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, dep.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, dep.Name, found.Name)
}

func TestDependencyRepository_GetByOwnerRepo(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	dep := &entity.Dependency{
		ID:    uuid.New(),
		Name:  "test-dependency",
		Owner: "testowner",
		Repo:  "testrepo",
	}
	err := repo.Create(ctx, dep)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByOwnerRepo(ctx, "testowner", "testrepo")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, dep.ID, found.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetByOwnerRepo(ctx, "nonexistent", "repo")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestDependencyRepository_GetByOwnerRepoCI(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	dep := &entity.Dependency{
		ID:    uuid.New(),
		Name:  "test-dependency",
		Owner: "TestOwner",
		Repo:  "TestRepo",
	}
	err := repo.Create(ctx, dep)
	require.NoError(t, err)

	t.Run("CaseInsensitiveMatch", func(t *testing.T) {
		found, err := repo.GetByOwnerRepoCI(ctx, "testowner", "testrepo")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, dep.ID, found.ID)
	})

	t.Run("ExactCaseMatch", func(t *testing.T) {
		found, err := repo.GetByOwnerRepoCI(ctx, "TestOwner", "TestRepo")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, dep.ID, found.ID)
	})
}

func TestDependencyRepository_SearchByName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	deps := []*entity.Dependency{
		{ID: uuid.New(), Name: "express", Owner: "owner1", Repo: "repo1"},
		{ID: uuid.New(), Name: "express-validator", Owner: "owner2", Repo: "repo2"},
		{ID: uuid.New(), Name: "lodash", Owner: "owner3", Repo: "repo3"},
	}

	for _, dep := range deps {
		err := repo.Create(ctx, dep)
		require.NoError(t, err)
	}

	results, err := repo.SearchByName(ctx, "express")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestDependencyRepository_GetByNameCI(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	dep := &entity.Dependency{
		ID:    uuid.New(),
		Name:  "TestPackage",
		Owner: "owner",
		Repo:  "repo",
	}
	err := repo.Create(ctx, dep)
	require.NoError(t, err)

	t.Run("CaseInsensitiveMatch", func(t *testing.T) {
		found, err := repo.GetByNameCI(ctx, "testpackage")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, dep.ID, found.ID)
	})

	t.Run("ExactCaseMatch", func(t *testing.T) {
		found, err := repo.GetByNameCI(ctx, "TestPackage")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, dep.ID, found.ID)
	})
}

func TestDependencyRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	dep := &entity.Dependency{
		ID:    uuid.New(),
		Name:  "test-dependency",
		Owner: "oldowner",
		Repo:  "oldrepo",
	}
	err := repo.Create(ctx, dep)
	require.NoError(t, err)

	dep.Owner = "newowner"
	dep.Repo = "newrepo"
	err = repo.Update(ctx, dep)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, "newowner", found.Owner)
	assert.Equal(t, "newrepo", found.Repo)
}

func TestDependencyRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	dep := &entity.Dependency{
		ID:    uuid.New(),
		Name:  "test-dependency",
		Owner: "owner",
		Repo:  "repo",
	}
	err := repo.Create(ctx, dep)
	require.NoError(t, err)

	err = repo.Delete(ctx, dep.ID)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, dep.ID)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestDependencyRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDependencyRepository(db)
	ctx := context.Background()

	deps := []*entity.Dependency{
		{ID: uuid.New(), Name: "dep1", Owner: "owner1", Repo: "repo1"},
		{ID: uuid.New(), Name: "dep2", Owner: "owner2", Repo: "repo2"},
		{ID: uuid.New(), Name: "dep3", Owner: "owner3", Repo: "repo3"},
	}

	for _, dep := range deps {
		err := repo.Create(ctx, dep)
		require.NoError(t, err)
	}

	results, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
}
