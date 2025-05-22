// internal/models/vendor.go
package models

import "time"

// Vendor represents a service provider (e.g., airline, hotel, tour operator).
type Vendor struct {
	ID            uint      `gorm:"primaryKey"`
	TenantID      uint      `gorm:"not null;index"` // Ensures vendor records are tenant-specific.
	Name          string    `gorm:"size:255;not null"`
	Type          string    `gorm:"size:100"`       // E.g., Airline, Hotel, Tour Operator, etc.
	ContactPerson string    `gorm:"size:255"`
	ContactInfo   string    `gorm:"size:255"`
	PaymentTerms  string    `gorm:"size:255"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
