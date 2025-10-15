package entity

import (
	"time"

	"github.com/google/uuid"
)

type AppDependency struct {
	ID                      uuid.UUID  `gorm:"primaryKey" db:"id" json:"id"`
	AppID                   uuid.UUID  `db:"app_id" json:"app_id"`
	DependencyID            uuid.UUID  `db:"dependency_id" json:"dependency_id"`
	UsedCommitSHA           *string    `db:"used_commit_sha" json:"used_commit_sha"`
	UsedVersion             string     `db:"used_version" json:"used_version"`
	UsedTag                 *string    `db:"used_tag" json:"used_tag"`
	IsMonitored             bool       `db:"is_monitored" json:"is_monitored"`
	MonitoringEnabled       bool       `db:"monitoring_enabled" json:"monitoring_enabled"`
	PollingIntervalMinutes  int        `db:"polling_interval_minutes" json:"polling_interval_minutes"`
	LastCheckedCommitSHA    *string    `db:"last_checked_commit_sha" json:"last_checked_commit_sha"`
	LastCheckedTag          *string    `db:"last_checked_tag" json:"last_checked_tag"`
	LastCheckedAt           *time.Time `db:"last_checked_at" json:"last_checked_at"`
	LastMonitoredAt         *time.Time `db:"last_monitored_at" json:"last_monitored_at"`
	MonitorStatus           *string    `db:"monitor_status" json:"monitor_status"`
	TotalChecksCount        int        `db:"total_checks_count" json:"total_checks_count"`
	LastSecurityDetectionAt *time.Time `db:"last_security_detection_at" json:"last_security_detection_at"`
	LastSecurityScore       int        `db:"last_security_score" json:"last_security_score"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

func (AppDependency) TableName() string {
	return "app_dependencies"
}
