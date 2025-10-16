package entity

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid" db:"id" json:"id"`
	Name        string    `gorm:"type:text;not null" db:"name" json:"name"`
	RuntimeID   *int      `gorm:"type:int" db:"runtime_id" json:"runtime_id"`
	FrameworkID *int      `gorm:"type:int" db:"framework_id" json:"framework_id"`
	Description *string   `gorm:"type:text" db:"description" json:"description"`
	IsDeleted   bool      `gorm:"not null;default:false" db:"is_deleted" json:"is_deleted"`
	Status      string    `gorm:"type:text" db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

func (App) TableName() string {
	return "app"
}
