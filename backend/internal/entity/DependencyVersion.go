package entity

import (
	"time"

	"github.com/google/uuid"
)

type DependencyVersion struct {
	ID           uuid.UUID `gorm:"primaryKey" db:"id" json:"id"`
	DependencyID uuid.UUID `db:"dependency_id" json:"dependency_id"`
	CommitSHA    string    `db:"commit_sha" json:"commit_sha"`
	CommitAt     time.Time `db:"commit_at" json:"commit_at"`
	Tag          *string   `db:"tag" json:"tag"`
	Branch       *string   `db:"branch" json:"branch"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

func (DependencyVersion) TableName() string {
	return "dependency_versions"
}
