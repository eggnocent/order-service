package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderField struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	OrderID         uint      `gorm:"type:bigint;not null"`
	FieldScheduleID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}
