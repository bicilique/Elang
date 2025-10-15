package entity

import (
	"time"

	"github.com/google/uuid"
)

type AuditTrail struct {
	ID uuid.UUID `gorm:"primaryKey" db:"id" json:"id"`

	EntityType string    `db:"entity_type" json:"entity_type"` // app, dependency, monitoring_job, etc.
	EntityID   uuid.UUID `db:"entity_id" json:"entity_id"`
	Action     string    `db:"action" json:"action"` // created, updated, deleted, monitored, etc.

	OldValues []byte `gorm:"type:jsonb" db:"old_values" json:"old_values"`
	NewValues []byte `gorm:"type:jsonb" db:"new_values" json:"new_values"`

	PerformedBy string    `db:"performed_by" json:"performed_by"`
	PerformedAt time.Time `db:"performed_at" json:"performed_at"`
	Context     []byte    `gorm:"type:jsonb" db:"context" json:"context"`

	// Security-specific fields
	SecurityRelevant bool    `db:"security_relevant" json:"security_relevant"`
	RiskLevel        *string `db:"risk_level" json:"risk_level"`
}

func (AuditTrail) TableName() string {
	return "audit_trail"
}
