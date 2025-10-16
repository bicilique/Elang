package entity

import (
	"time"

	"github.com/google/uuid"
)

type AppDependency struct {
	ID                      uuid.UUID   `gorm:"primaryKey;type:uuid" db:"id" json:"id"`
	AppID                   uuid.UUID   `gorm:"type:uuid;not null" db:"app_id" json:"app_id"`
	App                     *App        `gorm:"foreignKey:AppID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	DependencyID            uuid.UUID   `gorm:"type:uuid;not null" db:"dependency_id" json:"dependency_id"`
	Dependency              *Dependency `gorm:"foreignKey:DependencyID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	UsedCommitSHA           *string     `gorm:"type:varchar(64)" db:"used_commit_sha" json:"used_commit_sha"`
	UsedVersion             string      `gorm:"type:varchar(128);not null" db:"used_version" json:"used_version"`
	UsedTag                 *string     `gorm:"type:varchar(128)" db:"used_tag" json:"used_tag"`
	IsMonitored             bool        `gorm:"not null;default:false" db:"is_monitored" json:"is_monitored"`
	MonitoringEnabled       bool        `gorm:"not null;default:true" db:"monitoring_enabled" json:"monitoring_enabled"`
	PollingIntervalMinutes  int         `gorm:"not null;default:60" db:"polling_interval_minutes" json:"polling_interval_minutes"`
	LastCheckedCommitSHA    *string     `gorm:"type:text" db:"last_checked_commit_sha" json:"last_checked_commit_sha"`
	LastCheckedTag          *string     `gorm:"type:text" db:"last_checked_tag" json:"last_checked_tag"`
	LastCheckedAt           *time.Time  `db:"last_checked_at" json:"last_checked_at"`
	LastMonitoredAt         *time.Time  `db:"last_monitored_at" json:"last_monitored_at"`
	MonitorStatus           *string     `gorm:"type:varchar(32);default:'ready'" db:"monitor_status" json:"monitor_status"`
	TotalChecksCount        int         `gorm:"default:0" db:"total_checks_count" json:"total_checks_count"`
	LastSecurityDetectionAt *time.Time  `db:"last_security_detection_at" json:"last_security_detection_at"`
	LastSecurityScore       int         `gorm:"default:0" db:"last_security_score" json:"last_security_score"`
	CreatedAt               time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time   `db:"updated_at" json:"updated_at"`
}

func (AppDependency) TableName() string {
	return "app_dependencies"
}
