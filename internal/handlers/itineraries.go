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

type ItineraryHandler struct {
	DB *gorm.DB
}

func NewItineraryHandler(db *gorm.DB) *ItineraryHandler {
	return &ItineraryHandler{DB: db}
}

// CreateItinerary accepts a payload with both itinerary and its items,
// enforces tenant scope, and wraps in a single transaction.
func (h *ItineraryHandler) CreateItinerary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	// we expect camelCase JSON from the client
	var payload struct {
		Name       string                 `json:"name"`
		StartDate  string                 `json:"startDate"`
		EndDate    string                 `json:"endDate"`
		Status     string                 `json:"status"`
		TotalPrice float64                `json:"totalPrice"`
		Items      []models.ItineraryItem `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Parse dates from string to time.Time, assuming layout "2006-01-02"
	startDate, err := time.Parse("2006-01-02", payload.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", payload.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	// Build parent record
	itin := models.Itinerary{
		TenantID:   claims.TenantID,
		Name:       payload.Name,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     payload.Status,
		TotalPrice: payload.TotalPrice,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Run in transaction: insert parent + items
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&itin).Error; err != nil {
			return err
		}
		for i := range payload.Items {
			item := payload.Items[i]
			item.ItineraryID = itin.ID
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		http.Error(w, "Failed to create itinerary and items", http.StatusInternalServerError)
		return
	}

	// Reload with items
	if err := h.DB.Preload("Items").First(&itin, itin.ID).Error; err != nil {
		http.Error(w, "Failed to load created itinerary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itin)
}

// UpdateItinerary updates both parent and child items in one transaction.
func (h *ItineraryHandler) UpdateItinerary(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.Atoi(chi.URLParam(r, "itineraryID"))
	if err != nil {
		http.Error(w, "Invalid itinerary ID", http.StatusBadRequest)
		return
	}

	// Decode update payload
	var payload struct {
		Name       string                 `json:"name"`
		StartDate  time.Time              `json:"startDate"`
		EndDate    time.Time              `json:"endDate"`
		Status     string                 `json:"status"`
		TotalPrice float64                `json:"totalPrice"`
		Items      []models.ItineraryItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var itin models.Itinerary
	if err := h.DB.Preload("Items").First(&itin, id64).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Itinerary not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Transaction: update parent, delete old items, create new ones
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		itin.Name = payload.Name
		itin.StartDate = payload.StartDate
		itin.EndDate = payload.EndDate
		itin.Status = payload.Status
		itin.TotalPrice = payload.TotalPrice
		itin.UpdatedAt = time.Now()
		if err := tx.Save(&itin).Error; err != nil {
			return err
		}

		// remove old items
		if err := tx.Where("itinerary_id = ?", itin.ID).Delete(&models.ItineraryItem{}).Error; err != nil {
			return err
		}

		// insert new items
		for i := range payload.Items {
			item := payload.Items[i]
			item.ItineraryID = itin.ID
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		http.Error(w, "Failed to update itinerary and items", http.StatusInternalServerError)
		return
	}

	// Reload and return
	if err := h.DB.Preload("Items").First(&itin, itin.ID).Error; err != nil {
		http.Error(w, "Failed to load updated itinerary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itin)
}

// ListItineraries returns all itineraries (with items) for the current tenant.
func (h *ItineraryHandler) ListItineraries(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant information", http.StatusUnauthorized)
		return
	}

	var list []models.Itinerary
	if err := h.DB.
		Where("tenant_id = ?", claims.TenantID).
		Preload("Items").
		Find(&list).Error; err != nil {
		http.Error(w, "Failed to fetch itineraries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GetItinerary returns a single itinerary (with items), checking tenant ownership.
func (h *ItineraryHandler) GetItinerary(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.Atoi(chi.URLParam(r, "itineraryID"))
	if err != nil {
		http.Error(w, "Invalid itinerary ID", http.StatusBadRequest)
		return
	}

	var itin models.Itinerary
	if err := h.DB.Preload("Items").First(&itin, id64).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Itinerary not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok || itin.TenantID != claims.TenantID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itin)
}

// DeleteItinerary removes an itinerary and all its items in one transaction.
func (h *ItineraryHandler) DeleteItinerary(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.Atoi(chi.URLParam(r, "itineraryID"))
	if err != nil {
		http.Error(w, "Invalid itinerary ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("itinerary_id = ?", id64).Delete(&models.ItineraryItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Itinerary{}, id64).Error
	}); err != nil {
		http.Error(w, "Failed to delete itinerary", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
