// handlers/booking.go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"travel-agency/internal/auth"
	"travel-agency/internal/models"
)

type BookingHandler struct {
	DB *gorm.DB
}

func NewBookingHandler(db *gorm.DB) *BookingHandler {
	return &BookingHandler{DB: db}
}

// createBookingInput defines the fields clients may submit when creating.
type createBookingInput struct {
	ItineraryID uint      `json:"itineraryID"`
	VendorID    uint      `json:"vendorID"`
	BookingRef  string    `json:"bookingRef"`
	Status      string    `json:"status"`
	BookingDate time.Time `json:"bookingDate"`
	TravelDate  time.Time `json:"travelDate"`
	Cost        float64   `json:"cost"`
	Price       float64   `json:"price"`
}

// updateBookingInput defines the fields clients may submit when updating.
type updateBookingInput struct {
	BookingRef  *string    `json:"bookingRef,omitempty"`
	Status      *string    `json:"status,omitempty"`
	BookingDate *time.Time `json:"bookingDate,omitempty"`
	TravelDate  *time.Time `json:"travelDate,omitempty"`
	Cost        *float64   `json:"cost,omitempty"`
	Price       *float64   `json:"price,omitempty"`
}

// CreateBooking handles POST /bookings
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant info", http.StatusUnauthorized)
		return
	}

	var input createBookingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	booking := models.Booking{
		ItineraryID: input.ItineraryID,
		VendorID:    input.VendorID,
		BookingRef:  input.BookingRef,
		Status:      input.Status,
		BookingDate: input.BookingDate,
		TravelDate:  input.TravelDate,
		Cost:        input.Cost,
		Price:       input.Price,

		TenantID:  claims.TenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.DB.Create(&booking).Error; err != nil {
		http.Error(w, "Failed to create booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// ListBookings handles GET /bookings
func (h *BookingHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant info", http.StatusUnauthorized)
		return
	}

	var bookings []models.Booking
	if err := h.DB.
		Where("tenant_id = ?", claims.TenantID).
		Find(&bookings).Error; err != nil {
		http.Error(w, "Failed to fetch bookings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

// GetBooking handles GET /bookings/{bookingID}
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant info", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "bookingID"))
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var booking models.Booking
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Booking not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// UpdateBooking handles PUT /bookings/{bookingID}
func (h *BookingHandler) UpdateBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant info", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "bookingID"))
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var booking models.Booking
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Booking not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	var input updateBookingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Only update fields the user provided
	if input.BookingRef != nil {
		booking.BookingRef = *input.BookingRef
	}
	if input.Status != nil {
		booking.Status = *input.Status
	}
	if input.BookingDate != nil {
		booking.BookingDate = *input.BookingDate
	}
	if input.TravelDate != nil {
		booking.TravelDate = *input.TravelDate
	}
	if input.Cost != nil {
		booking.Cost = *input.Cost
	}
	if input.Price != nil {
		booking.Price = *input.Price
	}
	booking.UpdatedAt = time.Now()

	if err := h.DB.Save(&booking).Error; err != nil {
		http.Error(w, "Failed to update booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// DeleteBooking handles DELETE /bookings/{bookingID}
func (h *BookingHandler) DeleteBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Unauthorized: missing tenant info", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "bookingID"))
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		Delete(&models.Booking{}).Error; err != nil {
		http.Error(w, "Failed to delete booking", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
