// internal/models/booking.go
package models

import "time"

// Booking represents a travel reservation or booking.
type Booking struct {
	ID          uint      `gorm:"primaryKey"`
	TenantID    uint      `gorm:"not null;index"`            // Multi-tenant: associates booking with an agency.
	ItineraryID uint      // Optional: if the booking is part of a larger itinerary.
	VendorID    uint      `gorm:"not null;index"`            // References the vendor providing the service.
	BookingRef  string    `gorm:"size:255"`                  // Supplier confirmation code or PNR.
	Status      string    `gorm:"size:50;default:'Pending'"` // e.g., Pending, Confirmed, Canceled.
	BookingDate time.Time // The date the booking is created.
	TravelDate  time.Time // The travel date (or start date for hotels).
	Cost        float64   `gorm:"default:0"`               // The cost charged by the vendor.
	Price       float64   `gorm:"default:0"`               // The price charged to the client.
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
