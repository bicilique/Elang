package repository

import (
	"context"
	"elang-backend/internal/entity"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dependencyRepository struct {
	db *gorm.DB
}

func NewDependencyRepository(db *gorm.DB) DependencyRepository {
	return &dependencyRepository{db: db}
}

func (r *dependencyRepository) Create(ctx context.Context, dep *entity.Dependency) error {
	return r.db.WithContext(ctx).Create(dep).Error
}

func (r *dependencyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Dependency, error) {
	var dep entity.Dependency
	err := r.db.WithContext(ctx).First(&dep, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func (r *dependencyRepository) GetByOwnerRepo(ctx context.Context, owner, repo string) (*entity.Dependency, error) {
	var dep entity.Dependency
	err := r.db.WithContext(ctx).Where("owner = ? AND repo = ?", owner, repo).First(&dep).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func (r *dependencyRepository) GetByOwnerRepoCI(ctx context.Context, owner, repo string) (*entity.Dependency, error) {
	var dep entity.Dependency
	err := r.db.WithContext(ctx).
		Where("LOWER(owner) = ? AND LOWER(repo) = ?", strings.ToLower(owner), strings.ToLower(repo)).
		First(&dep).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func (r *dependencyRepository) GetAll(ctx context.Context) ([]*entity.Dependency, error) {
	var result []*entity.Dependency
	err := r.db.WithContext(ctx).Find(&result).Error
	return result, err
}

func (r *dependencyRepository) Update(ctx context.Context, dep *entity.Dependency) error {
	return r.db.WithContext(ctx).Save(dep).Error
}

func (r *dependencyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Dependency{}, "id = ?", id).Error
}

func (r *dependencyRepository) SearchByName(ctx context.Context, name string) ([]*entity.Dependency, error) {
	var result []*entity.Dependency

	// Use database-agnostic LIKE operation
	// SQLite uses LIKE (case-insensitive by default), PostgreSQL uses ILIKE
	dialectName := r.db.Dialector.Name()
	var query string
	if dialectName == "postgres" {
		query = "name ILIKE ?"
	} else {
		query = "name LIKE ?"
	}

	err := r.db.WithContext(ctx).Where(query, "%"+name+"%").Find(&result).Error
	return result, err
}

func (r *dependencyRepository) GetByNameCI(ctx context.Context, name string) (*entity.Dependency, error) {
	var dep entity.Dependency
	err := r.db.WithContext(ctx).Where("LOWER(name) = ?", strings.ToLower(name)).First(&dep).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dep, nil
}
