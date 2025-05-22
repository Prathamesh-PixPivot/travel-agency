// internal/models/notification.go
package models

import "time"

type Notification struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null;index"`
	UserID    uint   `gorm:"not null"` // The recipient.
	Title     string `gorm:"size:255;not null"`
	Message   string `gorm:"size:1024"`
	Read      bool   `gorm:"default:false"`
	CreatedAt time.Time
}
