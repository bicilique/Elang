package repository

import (
	"context"
	"elang-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dependencyVersionRepository struct {
	db *gorm.DB
}

func NewDependencyVersionRepository(db *gorm.DB) DependencyVersionRepository {
	return &dependencyVersionRepository{db: db}
}

func (r *dependencyVersionRepository) Create(ctx context.Context, ver *entity.DependencyVersion) error {
	return r.db.WithContext(ctx).Create(ver).Error
}

func (r *dependencyVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.DependencyVersion, error) {
	var ver entity.DependencyVersion
	err := r.db.WithContext(ctx).First(&ver, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

func (r *dependencyVersionRepository) GetByDependencyID(ctx context.Context, depID uuid.UUID) ([]*entity.DependencyVersion, error) {
	var result []*entity.DependencyVersion
	err := r.db.WithContext(ctx).Where("dependency_id = ?", depID).Order("commit_at DESC").Find(&result).Error
	return result, err
}

func (r *dependencyVersionRepository) Update(ctx context.Context, ver *entity.DependencyVersion) error {
	return r.db.WithContext(ctx).Save(ver).Error
}

func (r *dependencyVersionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.DependencyVersion{}, "id = ?", id).Error
}

func (r *dependencyVersionRepository) GetByTag(ctx context.Context, tag string) ([]*entity.DependencyVersion, error) {
	var result []*entity.DependencyVersion
	err := r.db.WithContext(ctx).Where("tag = ?", tag).Find(&result).Error
	return result, err
}
