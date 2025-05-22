package models

import "time"

type Lead struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    TenantID     uint      `json:"tenant_id"`
    CustomerName string    `json:"name"`           // Maps incoming "name" to CustomerName
    ContactInfo  string    `json:"email"`          // Maps incoming "email" to ContactInfo
    Phone        string    `json:"phone"`
    Destination  string    `json:"destination"`
    Budget       float64   `json:"budget"`
    TravelDate   time.Time `json:"travelDate"`     // Ensure your frontend sends a date string parseable to time.Time
    Details      string    `json:"notes"`          // Maps incoming "notes" to Details
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
    AssignedTo   uint      `json:"assignedTo"`     // Typically set from the admin/agent claims
}
