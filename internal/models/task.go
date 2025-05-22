// internal/models/task.go
package models

import "time"

// Task represents an internal task assigned to an agent.
type Task struct {
  ID          uint      `gorm:"primaryKey" json:"id"`
  TenantID    uint      `gorm:"not null;index" json:"tenantId"`
  Title       string    `gorm:"size:255;not null" json:"title"`
  Description string    `gorm:"size:1024" json:"description"`
  AssignedTo  uint      `json:"assignedTo"`                      // user ID
  Priority    string    `gorm:"size:50;default:'Normal'" json:"priority"` // Low, Normal, High
  Status      string    `gorm:"size:50;default:'Pending'" json:"status"` // Pending, In Progress, Completed
  DueDate     time.Time `json:"dueDate"`                             // client should send ISO8601
  CreatedAt   time.Time `json:"createdAt"`
  UpdatedAt   time.Time `json:"updatedAt"`
}
