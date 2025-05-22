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

type VendorHandler struct {
	DB *gorm.DB
}

func NewVendorHandler(db *gorm.DB) *VendorHandler {
	return &VendorHandler{DB: db}
}

func (h *VendorHandler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	var vendor models.Vendor
	if err := json.NewDecoder(r.Body).Decode(&vendor); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Force TenantID.
	vendor.TenantID = claims.TenantID
	vendor.CreatedAt = time.Now()
	vendor.UpdatedAt = time.Now()

	if err := h.DB.Create(&vendor).Error; err != nil {
		http.Error(w, "Failed to create vendor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vendor)
}

func (h *VendorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		http.Error(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	var vendors []models.Vendor
	if err := h.DB.Where("tenant_id = ?", claims.TenantID).Find(&vendors).Error; err != nil {
		http.Error(w, "Unable to fetch vendors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vendors)
}

func (h *VendorHandler) GetVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := strconv.Atoi(chi.URLParam(r, "vendorID"))
	if err != nil {
		http.Error(w, "Invalid vendor ID", http.StatusBadRequest)
		return
	}

	var vendor models.Vendor
	if err := h.DB.First(&vendor, vendorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Vendor not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vendor)
}

func (h *VendorHandler) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := strconv.Atoi(chi.URLParam(r, "vendorID"))
	if err != nil {
		http.Error(w, "Invalid vendor ID", http.StatusBadRequest)
		return
	}

	var vendor models.Vendor
	if err := h.DB.First(&vendor, vendorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Vendor not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var updated models.Vendor
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Update allowed fields.
	vendor.Name = updated.Name
	vendor.Type = updated.Type
	vendor.ContactPerson = updated.ContactPerson
	vendor.ContactInfo = updated.ContactInfo
	vendor.PaymentTerms = updated.PaymentTerms
	vendor.UpdatedAt = time.Now()

	if err := h.DB.Save(&vendor).Error; err != nil {
		http.Error(w, "Unable to update vendor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vendor)
}
