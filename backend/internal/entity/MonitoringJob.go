package entity

import (
	"time"

	"github.com/google/uuid"
)

// MonitoringJob represents active monitoring processes
type MonitoringJob struct {
	ID          uuid.UUID  `gorm:"primaryKey" db:"id" json:"id"`
	JobType     string     `db:"job_type" json:"job_type"` // scheduled, manual, on_demand
	Status      string     `db:"status" json:"status"`     // running, completed, failed, cancelled
	StartedAt   time.Time  `db:"started_at" json:"started_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at"`

	// Job configuration
	AppIDs                 []uuid.UUID `gorm:"type:text;serializer:json" db:"app_ids" json:"app_ids"`
	DependencyIDs          []uuid.UUID `gorm:"type:text;serializer:json" db:"dependency_ids" json:"dependency_ids"`
	PollingIntervalMinutes int         `db:"polling_interval_minutes" json:"polling_interval_minutes"`
	MaxConcurrentChecks    int         `db:"max_concurrent_checks" json:"max_concurrent_checks"`

	// Progress tracking
	TotalChecksPlanned int `db:"total_checks_planned" json:"total_checks_planned"`
	ChecksCompleted    int `db:"checks_completed" json:"checks_completed"`
	ChecksFailed       int `db:"checks_failed" json:"checks_failed"`
	SecurityDetections int `db:"security_detections" json:"security_detections"`

	// Results and errors
	ResultsSummary []byte  `gorm:"type:jsonb" db:"results_summary" json:"results_summary"`
	ErrorLog       *string `db:"error_log" json:"error_log"`

	CreatedBy string    `db:"created_by" json:"created_by"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
