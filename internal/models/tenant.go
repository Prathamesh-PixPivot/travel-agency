// internal/models/tenant.go
package models

import "time"

type Tenant struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Address   string    `gorm:"size:512"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
