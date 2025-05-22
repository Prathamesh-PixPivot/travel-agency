// internal/models/itinerary.go
package models

import "time"

type Itinerary struct {
    ID         uint             `gorm:"primaryKey"`
    TenantID   uint             `gorm:"not null;index"`
    CustomerID uint
    Name       string           `gorm:"size:255;not null"`
    StartDate  time.Time        `gorm:"not null"`
    EndDate    time.Time        `gorm:"not null"`
    Status     string           `gorm:"size:50;default:'Planned'"`
    TotalPrice float64          `gorm:"default:0"`
    CreatedAt  time.Time
    UpdatedAt  time.Time

    // Add this:
    Items      []ItineraryItem  `gorm:"foreignKey:ItineraryID" json:"items"`
}

type ItineraryItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ItineraryID uint      `gorm:"not null;index" json:"itinerary_id"`
	Day         int       `gorm:"not null" json:"day"`
	Type        string    `gorm:"size:50;not null" json:"type"`
	Description string    `gorm:"size:1024" json:"description"`
	VendorID    uint      `json:"vendor_id"`
	Cost        float64   `gorm:"default:0" json:"cost"`
	Price       float64   `gorm:"default:0" json:"price"`
	Status      string    `gorm:"size:50;default:'Pending'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
