package repository

import (
	"context"
	"elang-backend/internal/entity"
	"strings"

	"gorm.io/gorm"
)

type frameworkRepository struct {
	db *gorm.DB
}

func NewFrameworkRepository(db *gorm.DB) FrameworkRepository {
	return &frameworkRepository{db: db}
}

func (r *frameworkRepository) Create(ctx context.Context, framework *entity.Framework) error {
	return r.db.WithContext(ctx).Create(framework).Error
}

func (r *frameworkRepository) GetByID(ctx context.Context, id int) (*entity.Framework, error) {
	var fw entity.Framework
	err := r.db.WithContext(ctx).First(&fw, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fw, nil
}

func (r *frameworkRepository) GetAll(ctx context.Context) ([]*entity.Framework, error) {
	var result []*entity.Framework
	err := r.db.WithContext(ctx).Find(&result).Error
	return result, err
}

func (r *frameworkRepository) Update(ctx context.Context, framework *entity.Framework) error {
	return r.db.WithContext(ctx).Save(framework).Error
}

func (r *frameworkRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.Framework{}, "id = ?", id).Error
}

func (r *frameworkRepository) GetByName(ctx context.Context, name string) (*entity.Framework, error) {
	var fw entity.Framework
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&fw).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fw, nil
}

func (r *frameworkRepository) GetByNameCI(ctx context.Context, name string) (*entity.Framework, error) {
	var fw entity.Framework
	err := r.db.WithContext(ctx).Where("LOWER(name) = ?", strings.ToLower(name)).First(&fw).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fw, nil
}
