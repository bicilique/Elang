package entity

import (
	"time"

	"github.com/google/uuid"
)

type Dependency struct {
	ID            uuid.UUID  `gorm:"primaryKey;type:uuid" db:"id" json:"id"`
	Name          string     `gorm:"type:text;not null" db:"name" json:"name"`
	Owner         string     `gorm:"type:text;not null" db:"owner" json:"owner"`
	Repo          string     `gorm:"type:text;not null" db:"repo" json:"repo"`
	LastCommitSHA *string    `gorm:"type:text" db:"last_commit_sha" json:"last_commit_sha"`
	LastCommitAt  *time.Time `db:"last_commit_at" json:"last_commit_at"`
	LastTag       *string    `gorm:"type:text" db:"last_tag" json:"last_tag"`
	LastTagAt     *time.Time `db:"last_tag_at" json:"last_tag_at"`
	RepositoryURL *string    `gorm:"type:text" db:"repository_url" json:"repository_url"`
	DefaultBranch *string    `gorm:"type:text;default:'main'" db:"default_branch" json:"default_branch"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

func (Dependency) TableName() string {
	return "dependencies"
}
