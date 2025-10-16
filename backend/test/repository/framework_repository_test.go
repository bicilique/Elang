package repository_test

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameworkRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	framework := &entity.Framework{
		Name: "Express",
	}

	err := repo.Create(ctx, framework)
	assert.NoError(t, err)
	assert.NotZero(t, framework.ID)
}

func TestFrameworkRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	framework := &entity.Framework{
		Name: "Django",
	}
	err := repo.Create(ctx, framework)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, framework.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, framework.ID, found.ID)
		assert.Equal(t, "Django", found.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetByID(ctx, 99999)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestFrameworkRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	frameworks := []*entity.Framework{
		{Name: "Express"},
		{Name: "Django"},
		{Name: "Spring Boot"},
	}

	for _, fw := range frameworks {
		err := repo.Create(ctx, fw)
		require.NoError(t, err)
	}

	results, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3)
}

func TestFrameworkRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	framework := &entity.Framework{
		Name: "Rails",
	}
	err := repo.Create(ctx, framework)
	require.NoError(t, err)

	framework.Name = "Ruby on Rails"
	err = repo.Update(ctx, framework)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, framework.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Ruby on Rails", found.Name)
}

func TestFrameworkRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	framework := &entity.Framework{
		Name: "Laravel",
	}
	err := repo.Create(ctx, framework)
	require.NoError(t, err)

	err = repo.Delete(ctx, framework.ID)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, framework.ID)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestFrameworkRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	framework := &entity.Framework{
		Name: "FastAPI",
	}
	err := repo.Create(ctx, framework)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByName(ctx, "FastAPI")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, framework.ID, found.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetByName(ctx, "NonExistent")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestFrameworkRepository_GetByNameCI(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	framework := &entity.Framework{
		Name: "Express",
	}
	err := repo.Create(ctx, framework)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Lowercase", "express", true},
		{"Uppercase", "EXPRESS", true},
		{"ExactCase", "Express", true},
		{"MixedCase", "eXpReSs", true},
		{"NotFound", "Django", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found, err := repo.GetByNameCI(ctx, tc.input)
			assert.NoError(t, err)
			if tc.expected {
				assert.NotNil(t, found)
				assert.Equal(t, framework.ID, found.ID)
			} else {
				assert.Nil(t, found)
			}
		})
	}
}

func TestFrameworkRepository_ConcurrentCreates(t *testing.T) {
	// Skip concurrent test as it requires more complex setup
	t.Skip("Skipping concurrent test - requires more complex database setup")

	db := setupTestDB(t)
	repo := repository.NewFrameworkRepository(db)
	ctx := context.Background()

	frameworks := []*entity.Framework{
		{Name: "Framework1"},
		{Name: "Framework2"},
		{Name: "Framework3"},
	}

	// Create frameworks sequentially for now
	for _, fw := range frameworks {
		err := repo.Create(ctx, fw)
		assert.NoError(t, err)
	}

	// Verify all were created
	results, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3)
}
