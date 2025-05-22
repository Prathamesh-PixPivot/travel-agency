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

type TravelRequestHandler struct {
	DB *gorm.DB
}

func NewTravelRequestHandler(db *gorm.DB) *TravelRequestHandler {
	return &TravelRequestHandler{DB: db}
}

func (h *TravelRequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	var req models.TravelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	req.TenantID = claims.TenantID
	req.EmployeeID = claims.UserID
	req.Status = "Pending"
	req.SubmitDate = time.Now()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := h.DB.Create(&req).Error; err != nil {
		http.Error(w, "Failed to create travel request", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func (h *TravelRequestHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	requestID, err := strconv.Atoi(chi.URLParam(r, "requestID"))
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	var req models.TravelRequest
	if err := h.DB.First(&req, requestID).Error; err != nil {
		http.Error(w, "Travel request not found", http.StatusNotFound)
		return
	}

	// Here you might validate that the authenticated user is an approver.
	req.Status = "Approved"
	req.UpdatedAt = time.Now()

	if err := h.DB.Save(&req).Error; err != nil {
		http.Error(w, "Failed to update travel request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}
