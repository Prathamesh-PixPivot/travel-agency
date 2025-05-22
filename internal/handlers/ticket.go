// internal/handlers/ticket.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"travel-agency/internal/auth"
	"travel-agency/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate = validator.New()

type TicketHandler struct {
	DB *gorm.DB
}

func NewTicketHandler(db *gorm.DB) *TicketHandler {
	return &TicketHandler{DB: db}
}

// CreateTicket
func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)

	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := validate.Struct(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.(validator.ValidationErrors))
		return
	}

	ticket := models.Ticket{
		TenantID:    claims.TenantID,
		Subject:     req.Subject,
		Description: req.Description,
		CustomerID:  req.CustomerID,
		AssignedTo:  claims.UserID, // assign to creator by default
		Priority:    req.Priority,
	}

	if err := h.DB.Create(&ticket).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

// ListTickets
func (h *TicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)

	var tickets []models.Ticket
	if err := h.DB.
		Where("tenant_id = ?", claims.TenantID).
		Find(&tickets).
		Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

// GetTicket
func (h *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	ticketID, err := strconv.Atoi(chi.URLParam(r, "ticketID"))
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var ticket models.Ticket
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", ticketID, claims.TenantID).
		First(&ticket).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Not found", http.StatusNotFound)
		} else {
			http.Error(w, "DB error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// UpdateTicket
func (h *TicketHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	ticketID, err := strconv.Atoi(chi.URLParam(r, "ticketID"))
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := validate.Struct(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.(validator.ValidationErrors))
		return
	}

	// Build map of fields to update
	updates := make(map[string]interface{})
	if req.Subject != nil {
		updates["subject"] = *req.Subject
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.AssignedTo != nil {
		updates["assigned_to"] = *req.AssignedTo
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if err := h.DB.
		Model(&models.Ticket{}).
		Where("id = ? AND tenant_id = ?", ticketID, claims.TenantID).
		Updates(updates).
		Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteTicket
func (h *TicketHandler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	ticketID, err := strconv.Atoi(chi.URLParam(r, "ticketID"))
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.
		Where("id = ? AND tenant_id = ?", ticketID, claims.TenantID).
		Delete(&models.Ticket{}).
		Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
