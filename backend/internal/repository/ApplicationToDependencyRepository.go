package repository

import (
	"context"
	"elang-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type appDependencyRepository struct {
	db *gorm.DB
}

func NewAppDependencyRepository(db *gorm.DB) AppDependencyRepository {
	return &appDependencyRepository{db: db}
}

func (r *appDependencyRepository) Create(ctx context.Context, appDep *entity.AppDependency) error {
	return r.db.WithContext(ctx).Create(appDep).Error
}

func (r *appDependencyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AppDependency, error) {
	var appDep entity.AppDependency
	err := r.db.WithContext(ctx).First(&appDep, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &appDep, nil
}

func (r *appDependencyRepository) GetByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppDependency, error) {
	var result []*entity.AppDependency
	err := r.db.WithContext(ctx).Where("app_id = ?", appID).Find(&result).Error
	return result, err
}

func (r *appDependencyRepository) GetByDependencyID(ctx context.Context, depID uuid.UUID) ([]*entity.AppDependency, error) {
	var result []*entity.AppDependency
	err := r.db.WithContext(ctx).Where("dependency_id = ?", depID).Find(&result).Error
	return result, err
}

func (r *appDependencyRepository) Update(ctx context.Context, appDep *entity.AppDependency) error {
	return r.db.WithContext(ctx).Save(appDep).Error
}

func (r *appDependencyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.AppDependency{}, "id = ?", id).Error
}

func (r *appDependencyRepository) GetByStatus(ctx context.Context, status string) ([]*entity.AppDependency, error) {
	var result []*entity.AppDependency
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&result).Error
	return result, err
}

// GetByAppAndDependencyID fetches the AppDependency by app and dependency IDs
func (r *appDependencyRepository) GetByAppAndDependencyID(ctx context.Context, appID, depID uuid.UUID) (*entity.AppDependency, error) {
	var appDep entity.AppDependency
	err := r.db.WithContext(ctx).Where("app_id = ? AND dependency_id = ?", appID, depID).First(&appDep).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &appDep, nil
}
