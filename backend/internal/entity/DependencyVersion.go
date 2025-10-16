package entity

import (
	"time"

	"github.com/google/uuid"
)

type DependencyVersion struct {
	ID           uuid.UUID   `gorm:"primaryKey;type:uuid" db:"id" json:"id"`
	DependencyID uuid.UUID   `gorm:"type:uuid;not null" db:"dependency_id" json:"dependency_id"`
	Dependency   *Dependency `gorm:"foreignKey:DependencyID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CommitSHA    string      `gorm:"type:varchar(64);not null" db:"commit_sha" json:"commit_sha"`
	CommitAt     time.Time   `gorm:"not null" db:"commit_at" json:"commit_at"`
	Tag          *string     `gorm:"type:varchar(128)" db:"tag" json:"tag"`
	Branch       *string     `gorm:"type:text" db:"branch" json:"branch"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
}

func (DependencyVersion) TableName() string {
	return "dependency_versions"
}
