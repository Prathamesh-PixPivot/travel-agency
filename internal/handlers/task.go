// internal/handlers/task.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"travel-agency/internal/auth"
	"travel-agency/internal/models"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type TaskHandler struct {
	DB *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{DB: db}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant information", http.StatusUnauthorized)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Force TenantID from authenticated user.
	task.TenantID = claims.TenantID
	// If no AssignedTo specified, default to creator.
	if task.AssignedTo == 0 {
		task.AssignedTo = claims.UserID
	}

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	if err := h.DB.Create(&task).Error; err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []models.Task
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant info", http.StatusUnauthorized)
		return
	}
	if err := h.DB.Where("tenant_id = ?", claims.TenantID).Find(&tasks).Error; err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask returns one task by ID, scoped to the tenant.
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	if err := h.DB.First(&task, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if task.TenantID != claims.TenantID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask modifies an existing task.
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
	  http.Error(w, "Invalid task ID", http.StatusBadRequest)
	  return
	}
  
	var existing models.Task
	if err := h.DB.First(&existing, id).Error; err != nil {
	  if err == gorm.ErrRecordNotFound {
		http.Error(w, "Task not found", http.StatusNotFound)
	  } else {
		http.Error(w, "Database error", http.StatusInternalServerError)
	  }
	  return
	}
  
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if existing.TenantID != claims.TenantID {
	  http.Error(w, "Unauthorized", http.StatusUnauthorized)
	  return
	}
  
	var updated models.Task
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
	  http.Error(w, "Invalid payload", http.StatusBadRequest)
	  return
	}
  
	// Only allow mutating certain fields
	existing.Title = updated.Title
	existing.Description = updated.Description
	existing.AssignedTo = updated.AssignedTo
	existing.Priority = updated.Priority
	existing.Status = updated.Status
	existing.DueDate = updated.DueDate
	existing.UpdatedAt = time.Now()
  
	if err := h.DB.Save(&existing).Error; err != nil {
	  http.Error(w, "Failed to update task", http.StatusInternalServerError)
	  return
	}
  
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
  }
  
  // DeleteTask removes a task.
  func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
	  http.Error(w, "Invalid task ID", http.StatusBadRequest)
	  return
	}
  
	var task models.Task
	if err := h.DB.First(&task, id).Error; err != nil {
	  if err == gorm.ErrRecordNotFound {
		http.Error(w, "Task not found", http.StatusNotFound)
	  } else {
		http.Error(w, "Database error", http.StatusInternalServerError)
	  }
	  return
	}
  
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if task.TenantID != claims.TenantID {
	  http.Error(w, "Unauthorized", http.StatusUnauthorized)
	  return
	}
  
	if err := h.DB.Delete(&task).Error; err != nil {
	  http.Error(w, "Failed to delete task", http.StatusInternalServerError)
	  return
	}
  
	w.WriteHeader(http.StatusNoContent)
  }