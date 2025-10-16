package repository

import (
	"context"
	"elang-backend/internal/entity"
	"strings"

	"gorm.io/gorm"
)

type runtimeRepository struct {
	db *gorm.DB
}

func NewRuntimeRepository(db *gorm.DB) RuntimeRepository {
	return &runtimeRepository{db: db}
}

func (r *runtimeRepository) Create(ctx context.Context, runtime *entity.Runtime) error {
	return r.db.WithContext(ctx).Create(runtime).Error
}

func (r *runtimeRepository) GetByID(ctx context.Context, id int) (*entity.Runtime, error) {
	var rt entity.Runtime
	err := r.db.WithContext(ctx).First(&rt, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *runtimeRepository) GetAll(ctx context.Context) ([]*entity.Runtime, error) {
	var result []*entity.Runtime
	err := r.db.WithContext(ctx).Find(&result).Error
	return result, err
}

func (r *runtimeRepository) Update(ctx context.Context, runtime *entity.Runtime) error {
	return r.db.WithContext(ctx).Save(runtime).Error
}

func (r *runtimeRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&entity.Runtime{}, "id = ?", id).Error
}

func (r *runtimeRepository) GetByName(ctx context.Context, name string) (*entity.Runtime, error) {
	var rt entity.Runtime
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&rt).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *runtimeRepository) GetByNameCI(ctx context.Context, name string) (*entity.Runtime, error) {
	var rt entity.Runtime
	err := r.db.WithContext(ctx).Where("LOWER(name) = ?", strings.ToLower(name)).First(&rt).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}
