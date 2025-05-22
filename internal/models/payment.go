// internal/models/payment.go
package models

import "time"

// Payment represents a record of a payment made (or received) against an invoice.
type Payment struct {
	ID          uint      `gorm:"primaryKey"`
	TenantID    uint      `gorm:"not null;index"`
	InvoiceID   uint      `gorm:"not null;index"` // The invoice this payment is for.
	PaymentDate time.Time `gorm:"not null"`
	Amount      float64   `gorm:"not null"`
	Method      string    `gorm:"size:50"`                // e.g., "Credit Card", "Bank Transfer".
	Status      string    `gorm:"size:50;default:'Pending'"` // e.g., Pending, Completed, Failed.
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
