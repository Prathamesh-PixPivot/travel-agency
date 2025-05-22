// internal/models/audit_log.go
package models

import "time"

type AuditLog struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"index;not null"`
	UserID    uint   // The user performing the action.
	Action    string `gorm:"size:255;not null"`
	Entity    string `gorm:"size:255;not null"` // e.g., "Invoice", "Booking"
	EntityID  uint   //optional: ID of the entity being acted upon.
	Details   string `gorm:"size:1024"` // Optional: JSON describing changes.
	CreatedAt time.Time
}
