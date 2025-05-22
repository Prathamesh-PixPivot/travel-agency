// internal/jobs/auto_create_tasks.go
package jobs

import (
	"log"
	"time"

	"travel-agency/internal/models"

	"gorm.io/gorm"
)

// CreateFollowupTasks checks leads not contacted for more than 48 hours.
func CreateFollowupTasks(db *gorm.DB) {
	var leads []models.Lead
	cutoff := time.Now().Add(-48 * time.Hour)
	if err := db.Where("created_at < ? AND status = ?", cutoff, "New").Find(&leads).Error; err != nil {
		log.Printf("Error fetching stale leads: %v", err)
		return
	}

	for _, lead := range leads {
		// Create a task for follow-up.
		task := models.Task{
			TenantID:    lead.TenantID,
			Title:       "Follow-up Lead: " + lead.CustomerName,
			Description: "Please contact this lead as soon as possible.",
			Priority:    "High",
			Status:      "Pending",
			DueDate:     time.Now().Add(4 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := db.Create(&task).Error; err != nil {
			log.Printf("Failed to create follow-up task for lead %d: %v", lead.ID, err)
		}
	}
}
