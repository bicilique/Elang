package entity

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	ID          uuid.UUID `gorm:"primaryKey" db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	RuntimeID   *int      `db:"runtime_id" json:"runtime_id"`
	FrameworkID *int      `db:"framework_id" json:"framework_id"`
	Description *string   `db:"description" json:"description"`
	IsDeleted   bool      `db:"is_deleted" json:"is_deleted"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

func (App) TableName() string {
	return "app"
}
