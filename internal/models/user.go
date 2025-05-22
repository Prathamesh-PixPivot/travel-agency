// internal/models/user.go
package models

import "time"

type User struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	TenantID             uint      `gorm:"not null;index" json:"tenantId"`
	Name                 string    `gorm:"size:255;not null" json:"name"`
	Email                string    `gorm:"size:255;not null;uniqueIndex:idx_tenant_email" json:"email"`
	PasswordHash         string    `gorm:"not null" json:"-"`
	Role                 string    `gorm:"size:50;not null" json:"role"`
	IsActive             bool      `gorm:"default:true" json:"isActive"`
	ForcePasswordChange  bool      `gorm:"default:false" json:"forcePasswordChange"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}
