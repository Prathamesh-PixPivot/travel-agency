package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"travel-agency/internal/auth"
	"travel-agency/internal/models"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	DB *gorm.DB
}

func NewPaymentHandler(db *gorm.DB) *PaymentHandler {
	return &PaymentHandler{DB: db}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	var payment models.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	payment.TenantID = claims.TenantID
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()

	if err := h.DB.Create(&payment).Error; err != nil {
		http.Error(w, "Failed to create payment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	var payments []models.Payment
	if err := h.DB.Where("tenant_id = ?", claims.TenantID).Find(&payments).Error; err != nil {
		http.Error(w, "Unable to fetch payments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID, err := strconv.Atoi(chi.URLParam(r, "paymentID"))
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	var payment models.Payment
	if err := h.DB.First(&payment, paymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Payment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	paymentID, err := strconv.Atoi(chi.URLParam(r, "paymentID"))
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	var payment models.Payment
	if err := h.DB.First(&payment, paymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Payment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var updated models.Payment
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Update permitted fields only.
	payment.InvoiceID = updated.InvoiceID
	payment.PaymentDate = updated.PaymentDate
	payment.Amount = updated.Amount
	payment.Method = updated.Method
	payment.Status = updated.Status
	payment.UpdatedAt = time.Now()

	if err := h.DB.Save(&payment).Error; err != nil {
		http.Error(w, "Failed to update payment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}
