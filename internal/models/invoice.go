// internal/models/invoice.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Invoice represents a billing invoice (sale or purchase) for a given tenant.
type Invoice struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TenantID    uint       `gorm:"type:uint;not null;index" json:"tenantId"`
	InvoiceType string     `gorm:"size:20;not null" json:"invoiceType"` // "sale" or "purchase"
	IssueDate   time.Time  `gorm:"not null" json:"issueDate"`
	DueDate     time.Time  `gorm:"not null" json:"dueDate"`
	Status      string     `gorm:"size:50;not null;default:'Draft'" json:"status"`
	Amount      float64    `gorm:"not null;default:0" json:"amount"`
	Currency    string     `gorm:"size:3;not null;default:'USD'" json:"currency"`
	CustomerID  *uuid.UUID `gorm:"type:uuid;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"customerId,omitempty"`
	VendorID    *uuid.UUID `gorm:"type:uuid;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"vendorId,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// BeforeCreate hook to ensure ID is set to a new UUID, even if client supplies one.
func (inv *Invoice) BeforeCreate(tx *gorm.DB) (err error) {
	inv.ID = uuid.New()
	return
}
