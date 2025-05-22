// models/ticket.go
package models

import "time"

type Ticket struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    TenantID    uint      `gorm:"not null;index" json:"-"`
    Subject     string    `gorm:"size:255;not null" json:"subject"`
    Description string    `gorm:"size:1024" json:"description"`
    CustomerID  uint      `json:"customer_id"`
    AssignedTo  uint      `json:"assigned_to"`
    Status      string    `gorm:"size:50;default:Open" json:"status"`
    Priority    string    `gorm:"size:50;default:Normal" json:"priority"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// Use a separate struct for create/update payloads:
type CreateTicketRequest struct {
    Subject     string `json:"subject"     validate:"required,min=3"`
    Description string `json:"description"`
    CustomerID  uint   `json:"customer_id" validate:"required"`
    Priority    string `json:"priority"    validate:"oneof=Low Normal High"`
}

type UpdateTicketRequest struct {
    Subject     *string `json:"subject"     validate:"omitempty,min=3"`
    Description *string `json:"description"`
    AssignedTo  *uint   `json:"assigned_to"`
    Status      *string `json:"status"      validate:"omitempty,oneof=Open InProgress Closed"`
    Priority    *string `json:"priority"    validate:"omitempty,oneof=Low Normal High"`
}
