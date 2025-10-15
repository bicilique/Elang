package repository

import (
	"context"
	"elang-backend/internal/entity"
	"time"

	"github.com/google/uuid"
)

// Repository interfaces for each entity
type RuntimeRepository interface {
	Create(ctx context.Context, runtime *entity.Runtime) error
	GetByID(ctx context.Context, id int) (*entity.Runtime, error)
	GetAll(ctx context.Context) ([]*entity.Runtime, error)
	Update(ctx context.Context, runtime *entity.Runtime) error
	Delete(ctx context.Context, id int) error
	GetByName(ctx context.Context, name string) (*entity.Runtime, error)
	GetByNameCI(ctx context.Context, name string) (*entity.Runtime, error)
}

type FrameworkRepository interface {
	Create(ctx context.Context, framework *entity.Framework) error
	GetByID(ctx context.Context, id int) (*entity.Framework, error)
	GetAll(ctx context.Context) ([]*entity.Framework, error)
	Update(ctx context.Context, framework *entity.Framework) error
	Delete(ctx context.Context, id int) error
	GetByName(ctx context.Context, name string) (*entity.Framework, error)
	GetByNameCI(ctx context.Context, name string) (*entity.Framework, error)
}

type ApplicationRepository interface {
	Create(ctx context.Context, app *entity.App) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error)
	GetAll(ctx context.Context) ([]*entity.App, error)
	Update(ctx context.Context, app *entity.App) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (*entity.App, error)
	GetByStatus(ctx context.Context, status string) ([]*entity.App, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type DependencyRepository interface {
	Create(ctx context.Context, dep *entity.Dependency) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Dependency, error)
	GetByOwnerRepo(ctx context.Context, owner, repo string) (*entity.Dependency, error)
	GetAll(ctx context.Context) ([]*entity.Dependency, error)
	Update(ctx context.Context, dep *entity.Dependency) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByNameCI(ctx context.Context, name string) (*entity.Dependency, error)
	SearchByName(ctx context.Context, name string) ([]*entity.Dependency, error)
	GetByOwnerRepoCI(ctx context.Context, owner, repo string) (*entity.Dependency, error)
}

type AppDependencyRepository interface {
	Create(ctx context.Context, appDep *entity.AppDependency) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AppDependency, error)
	GetByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppDependency, error)
	GetByDependencyID(ctx context.Context, depID uuid.UUID) ([]*entity.AppDependency, error)
	Update(ctx context.Context, appDep *entity.AppDependency) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByStatus(ctx context.Context, status string) ([]*entity.AppDependency, error)
	GetByAppAndDependencyID(ctx context.Context, appID, depID uuid.UUID) (*entity.AppDependency, error)
}

type DependencyVersionRepository interface {
	Create(ctx context.Context, ver *entity.DependencyVersion) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.DependencyVersion, error)
	GetByDependencyID(ctx context.Context, depID uuid.UUID) ([]*entity.DependencyVersion, error)
	Update(ctx context.Context, ver *entity.DependencyVersion) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTag(ctx context.Context, tag string) ([]*entity.DependencyVersion, error)
}

type AuditTrailRepository interface {
	Create(ctx context.Context, audit *entity.AuditTrail) error
	LogAction(ctx context.Context, entityType string, entityID uuid.UUID, action string, oldValues, newValues interface{}, performedBy string) error
	LogSecurityEvent(ctx context.Context, entityType string, entityID uuid.UUID, action string, riskLevel string, context interface{}, performedBy string) error
	GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit, offset int) ([]*entity.AuditTrail, error)
	GetSecurityEvents(ctx context.Context, limit, offset int) ([]*entity.AuditTrail, error)
	GetByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*entity.AuditTrail, error)
	CleanupOldRecords(ctx context.Context, olderThan time.Time) error
}
