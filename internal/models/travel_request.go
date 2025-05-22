package models

import "time"

type TravelRequest struct {
	ID           uint      `gorm:"primaryKey"`
	TenantID     uint      `gorm:"not null;index"`
	EmployeeID   uint      // FK to User who submitted the request.
	TripDetails  string    `gorm:"size:1024"`
	Status       string    `gorm:"size:50;default:'Pending'"`
	SubmitDate   time.Time `gorm:"not null"`
	NeededByDate time.Time `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
