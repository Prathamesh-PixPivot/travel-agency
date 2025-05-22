package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"travel-agency/internal/auth"
	"travel-agency/internal/models"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type LeadsHandler struct {
	DB *gorm.DB
}

func NewLeadsHandler(db *gorm.DB) *LeadsHandler {
	return &LeadsHandler{DB: db}
}

func (h *LeadsHandler) CreateLead(w http.ResponseWriter, r *http.Request) {
	// Extract authenticated user's claims.
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant information", http.StatusUnauthorized)
		return
	}

	var lead models.Lead
	if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Force the TenantID from claims.
	lead.TenantID = claims.TenantID
	// Automatically assign the lead to the agent creating it.
	lead.AssignedTo = claims.UserID

	if err := h.DB.Create(&lead).Error; err != nil {
		http.Error(w, "Unable to create lead", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

func (h *LeadsHandler) ListLeads(w http.ResponseWriter, r *http.Request) {
	// Get TenantID from context.
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	var leads []models.Lead
	if err := h.DB.Where("tenant_id = ?", claims.TenantID).Find(&leads).Error; err != nil {
		http.Error(w, "Unable to fetch leads", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

func (h *LeadsHandler) GetLead(w http.ResponseWriter, r *http.Request) {
	leadID, err := strconv.Atoi(chi.URLParam(r, "leadID"))
	if err != nil {
		http.Error(w, "Invalid lead ID", http.StatusBadRequest)
		return
	}

	var lead models.Lead
	if err := h.DB.First(&lead, leadID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Lead not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Optionally: Verify that lead.TenantID matches the authenticated TenantID.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

func (h *LeadsHandler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	leadID, err := strconv.Atoi(chi.URLParam(r, "leadID"))
	if err != nil {
		http.Error(w, "Invalid lead ID", http.StatusBadRequest)
		return
	}

	var lead models.Lead
	if err := h.DB.First(&lead, leadID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Lead not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var updated models.Lead
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Update permitted fields; add others as needed:
	lead.CustomerName = updated.CustomerName
	lead.ContactInfo = updated.ContactInfo
	lead.Phone = updated.Phone
	lead.Destination = updated.Destination
	lead.Budget = updated.Budget
	// For TravelDate, ensure proper parsing or conversion.
	lead.TravelDate = updated.TravelDate
	lead.Details = updated.Details
	lead.Status = updated.Status
	lead.AssignedTo = updated.AssignedTo // Update if allowed

	lead.UpdatedAt = time.Now()

	if err := h.DB.Save(&lead).Error; err != nil {
		http.Error(w, "Unable to update lead", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}
