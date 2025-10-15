package repository

import (
	"context"
	"elang-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type appRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) ApplicationRepository {
	return &appRepository{db: db}
}

func (r *appRepository) Create(ctx context.Context, app *entity.App) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *appRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	var app entity.App
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) GetAll(ctx context.Context) ([]*entity.App, error) {
	var result []*entity.App
	err := r.db.WithContext(ctx).Find(&result).Error
	return result, err
}

func (r *appRepository) Update(ctx context.Context, app *entity.App) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *appRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.App{}, "id = ?", id).Error
}

func (r *appRepository) GetByName(ctx context.Context, name string) (*entity.App, error) {
	var app entity.App
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) GetByStatus(ctx context.Context, status string) ([]*entity.App, error) {
	var result []*entity.App
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&result).Error
	return result, err
}

// UpdateStatus updates only the status field of an app by ID.
func (r *appRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&entity.App{}).Where("id = ?", id).Update("status", status).Error
}
