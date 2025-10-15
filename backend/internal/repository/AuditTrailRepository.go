package repository

import (
	"context"
	"elang-backend/internal/entity"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditTrailRepository struct {
	db *gorm.DB
}

func NewAuditTrailRepository(db *gorm.DB) AuditTrailRepository {
	return &auditTrailRepository{db: db}
}

func (r *auditTrailRepository) Create(ctx context.Context, audit *entity.AuditTrail) error {
	return r.db.WithContext(ctx).Create(audit).Error
}

func (r *auditTrailRepository) LogAction(ctx context.Context, entityType string, entityID uuid.UUID, action string, oldValues, newValues interface{}, performedBy string) error {
	// Marshal oldValues and newValues to JSON bytes
	var oldValuesBytes, newValuesBytes []byte
	var err error

	if oldValues != nil {
		oldValuesBytes, err = json.Marshal(oldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old values: %w", err)
		}
	}

	if newValues != nil {
		newValuesBytes, err = json.Marshal(newValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new values: %w", err)
		}
	}

	audit := &entity.AuditTrail{
		EntityType:       entityType,
		EntityID:         entityID,
		Action:           action,
		OldValues:        oldValuesBytes,
		NewValues:        newValuesBytes,
		PerformedBy:      performedBy,
		PerformedAt:      time.Now(),
		SecurityRelevant: false,
	}
	return r.Create(ctx, audit)
}

func (r *auditTrailRepository) LogSecurityEvent(ctx context.Context, entityType string, entityID uuid.UUID, action string, riskLevel string, context interface{}, performedBy string) error {
	// Marshal context to JSON bytes
	var contextBytes []byte
	var err error

	if context != nil {
		contextBytes, err = json.Marshal(context)
		if err != nil {
			return fmt.Errorf("failed to marshal context: %w", err)
		}
	}

	audit := &entity.AuditTrail{
		EntityType:       entityType,
		EntityID:         entityID,
		Action:           action,
		Context:          contextBytes,
		PerformedBy:      performedBy,
		PerformedAt:      time.Now(),
		SecurityRelevant: true,
		RiskLevel:        &riskLevel,
	}
	return r.Create(ctx, audit)
}

func (r *auditTrailRepository) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit, offset int) ([]*entity.AuditTrail, error) {
	var audits []*entity.AuditTrail
	query := r.db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", entityType, entityID).Order("performed_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&audits).Error
	return audits, err
}

func (r *auditTrailRepository) GetSecurityEvents(ctx context.Context, limit, offset int) ([]*entity.AuditTrail, error) {
	var audits []*entity.AuditTrail
	query := r.db.WithContext(ctx).Where("security_relevant = true").Order("performed_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&audits).Error
	return audits, err
}

func (r *auditTrailRepository) GetByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*entity.AuditTrail, error) {
	var audits []*entity.AuditTrail
	query := r.db.WithContext(ctx).Where("performed_at BETWEEN ? AND ?", startTime, endTime).Order("performed_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&audits).Error
	return audits, err
}

func (r *auditTrailRepository) CleanupOldRecords(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).Where("performed_at < ? AND security_relevant = false", olderThan).Delete(&entity.AuditTrail{}).Error
}
