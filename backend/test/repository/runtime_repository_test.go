package repository_test

import (
	"context"
	"elang-backend/internal/entity"
	"elang-backend/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRuntimeRepository(db)
	ctx := context.Background()

	runtime := &entity.Runtime{
		Name: "Node.js",
	}

	err := repo.Create(ctx, runtime)
	assert.NoError(t, err)
	assert.NotZero(t, runtime.ID)

	found, err := repo.GetByID(ctx, runtime.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, runtime.Name, found.Name)
}

func TestRuntimeRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRuntimeRepository(db)
	ctx := context.Background()

	runtime := &entity.Runtime{
		Name: "Python",
	}
	err := repo.Create(ctx, runtime)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, runtime.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, runtime.ID, found.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetByID(ctx, 99999)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestRuntimeRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRuntimeRepository(db)
	ctx := context.Background()

	runtimes := []*entity.Runtime{
		{Name: "Node.js"},
		{Name: "Python"},
		{Name: "Go"},
	}

	for _, rt := range runtimes {
		err := repo.Create(ctx, rt)
		require.NoError(t, err)
	}

	results, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3)
}

func TestRuntimeRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRuntimeRepository(db)
	ctx := context.Background()

	runtime := &entity.Runtime{
		Name: "Node.js",
	}
	err := repo.Create(ctx, runtime)
	require.NoError(t, err)

	runtime.Name = "NodeJS"
	err = repo.Update(ctx, runtime)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, runtime.ID)
	assert.NoError(t, err)
	assert.Equal(t, "NodeJS", found.Name)
}

func TestRuntimeRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRuntimeRepository(db)
	ctx := context.Background()

	runtime := &entity.Runtime{
		Name: "Ruby",
	}
	err := repo.Create(ctx, runtime)
	require.NoError(t, err)

	err = repo.Delete(ctx, runtime.ID)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, runtime.ID)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestRuntimeRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRuntimeRepository(db)
	ctx := context.Background()

	runtime := &entity.Runtime{
		Name: "Java",
	}
	err := repo.Create(ctx, runtime)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByName(ctx, "Java")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, runtime.ID, found.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetByName(ctx, "NonExistent")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestRuntimeRepository_GetByNameCI(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRuntimeRepository(db)
	ctx := context.Background()

	runtime := &entity.Runtime{
		Name: "Node.js",
	}
	err := repo.Create(ctx, runtime)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Lowercase", "node.js", true},
		{"Uppercase", "NODE.JS", true},
		{"ExactCase", "Node.js", true},
		{"MixedCase", "nOdE.Js", true},
		{"NotFound", "python", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found, err := repo.GetByNameCI(ctx, tc.input)
			assert.NoError(t, err)
			if tc.expected {
				assert.NotNil(t, found)
				assert.Equal(t, runtime.ID, found.ID)
			} else {
				assert.Nil(t, found)
			}
		})
	}
}
