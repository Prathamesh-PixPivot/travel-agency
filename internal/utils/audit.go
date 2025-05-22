// internal/utils/audit.go
package utils

import (
	"time"

	"gorm.io/gorm"
	"travel-agency/internal/models"
)

// LogAction records an audit log entry.
func LogAction(db *gorm.DB, tenantID, userID uint, action, entity string, details string) error {
	audit := models.AuditLog{
		TenantID:  tenantID,
		UserID:    userID,
		Action:    action,
		Entity:    entity,
		Details:   details,
		CreatedAt: time.Now(),
	}
	return db.Create(&audit).Error
}
