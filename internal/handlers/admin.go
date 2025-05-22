// internal/handlers/admin.go
package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"travel-agency/internal/auth"
	"travel-agency/internal/models"
	"travel-agency/internal/notifications"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateAgentRequest represents the payload to create a new agent.
type CreateAgentRequest struct {
	Name  string `json:"name"`  // optional
	Email string `json:"email"` // required
	Role  string `json:"role"`  // defaults to "agent"
}

// AdminHandler holds dependencies for admin operations.
type AdminHandler struct {
	DB          *gorm.DB
	EmailSender notifications.EmailSender
}

// NewAdminHandler constructs an AdminHandler.
func NewAdminHandler(db *gorm.DB, sender notifications.EmailSender) *AdminHandler {
	return &AdminHandler{DB: db, EmailSender: sender}
}

// CreateAgent creates a new agent under the current tenant, sends a temp password.
func (h *AdminHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)

	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "agent"
	}

	// Generate temporary password
	tempPwd := generateTempPassword(12)
	hashed, err := bcrypt.GenerateFromPassword([]byte(tempPwd), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		TenantID:            claims.TenantID,
		Name:                req.Name,
		Email:               req.Email,
		PasswordHash:        string(hashed),
		Role:                req.Role,
		IsActive:            true,
		ForcePasswordChange: true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := h.DB.Create(user).Error; err != nil {
		http.Error(w, "Failed to create agent", http.StatusInternalServerError)
		return
	}

	// Send temporary password via email
	subject := "Your Temporary Password"
	body := "Hello " + req.Name + ",<br/><br/>" +
		"Your account has been created. Temporary password: <strong>" + tempPwd + "</strong><br/>" +
		"Please log in and change your password immediately."
	if err := h.EmailSender.SendEmail(req.Email, subject, body); err != nil {
		log.Printf("Warning: failed to send email to %s: %v", req.Email, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ListAgents returns all agents in the tenant.
func (h *AdminHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	var users []models.User
	if err := h.DB.
		Where("tenant_id = ?", claims.TenantID).
		Find(&users).Error; err != nil {
		http.Error(w, "Failed to list agents", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetAgent returns a single agent by ID.
func (h *AdminHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	id, err := strconv.Atoi(chi.URLParam(r, "agentID"))
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		First(&user).Error; err != nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateAgent modifies an existing agent's details.
func (h *AdminHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	id, err := strconv.Atoi(chi.URLParam(r, "agentID"))
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		First(&user).Error; err != nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	var payload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	user.Name = payload.Name
	user.Email = payload.Email
	user.Role = payload.Role
	user.UpdatedAt = time.Now()

	if err := h.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update agent", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteAgent removes an agent from the tenant.
func (h *AdminHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	id, err := strconv.Atoi(chi.URLParam(r, "agentID"))
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		Delete(&models.User{}).Error; err != nil {
		http.Error(w, "Failed to delete agent", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// generateTempPassword produces a random alphanumeric string of length n.
func generateTempPassword(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
