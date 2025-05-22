// handlers/invoice.go
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"travel-agency/internal/auth"
	"travel-agency/internal/models"
	"travel-agency/internal/utils"
)

// jsonError is a helper to send JSON‐encoded error messages.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type InvoiceHandler struct {
	DB *gorm.DB
}

func NewInvoiceHandler(db *gorm.DB) *InvoiceHandler {
	return &InvoiceHandler{DB: db}
}

// CreateInvoice handles POST /invoices
func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		jsonError(w, "Missing tenant information", http.StatusUnauthorized)
		return
	}

	var payload struct {
		InvoiceType string     `json:"invoiceType"`
		IssueDate   time.Time  `json:"issueDate"`
		DueDate     time.Time  `json:"dueDate"`
		Status      string     `json:"status"`
		Amount      float64    `json:"amount"`
		Currency    string     `json:"currency"`
		CustomerID  *uuid.UUID `json:"customerId,omitempty"`
		VendorID    *uuid.UUID `json:"vendorId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if payload.InvoiceType != "sale" && payload.InvoiceType != "purchase" {
		jsonError(w, "invoiceType must be 'sale' or 'purchase'", http.StatusBadRequest)
		return
	}
	if payload.IssueDate.After(payload.DueDate) {
		jsonError(w, "issueDate cannot be after dueDate", http.StatusBadRequest)
		return
	}
	if payload.Amount < 0 {
		jsonError(w, "amount must be non-negative", http.StatusBadRequest)
		return
	}

	invoice := models.Invoice{
		TenantID:    claims.TenantID,
		InvoiceType: payload.InvoiceType,
		IssueDate:   payload.IssueDate,
		DueDate:     payload.DueDate,
		Status:      payload.Status,
		Amount:      payload.Amount,
		Currency:    payload.Currency,
		CustomerID:  payload.CustomerID,
		VendorID:    payload.VendorID,
	}

	// Wrap creation + audit log in a transaction
	tx := h.DB.Begin()
	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		jsonError(w, "Failed to create invoice", http.StatusInternalServerError)
		return
	}
	_ = utils.LogAction(tx, claims.TenantID, claims.UserID, "CREATE_INVOICE", "Invoice", "Created invoice")
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(invoice)
}

// ListInvoices handles GET /invoices
func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		jsonError(w, "Unauthorized: missing tenant information", http.StatusUnauthorized)
		return
	}

	var invoices []models.Invoice
	if err := h.DB.
		Where("tenant_id = ?", claims.TenantID).
		Find(&invoices).Error; err != nil {
		jsonError(w, "Unable to fetch invoices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoices)
}

// GetInvoice handles GET /invoices/{invoiceID}
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "invoiceID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		jsonError(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var invoice models.Invoice
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Invoice not found", http.StatusNotFound)
		} else {
			jsonError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}

// UpdateInvoice handles PUT /invoices/{invoiceID}
func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "invoiceID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		jsonError(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch & verify tenant in one query
	var invoice models.Invoice
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Invoice not found", http.StatusNotFound)
		} else {
			jsonError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	var payload struct {
		InvoiceType string     `json:"invoiceType"`
		IssueDate   time.Time  `json:"issueDate"`
		DueDate     time.Time  `json:"dueDate"`
		Status      string     `json:"status"`
		Amount      float64    `json:"amount"`
		Currency    string     `json:"currency"`
		CustomerID  *uuid.UUID `json:"customerId,omitempty"`
		VendorID    *uuid.UUID `json:"vendorId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Validation (same as create)
	if payload.InvoiceType != "sale" && payload.InvoiceType != "purchase" {
		jsonError(w, "invoiceType must be 'sale' or 'purchase'", http.StatusBadRequest)
		return
	}
	if payload.IssueDate.After(payload.DueDate) {
		jsonError(w, "issueDate cannot be after dueDate", http.StatusBadRequest)
		return
	}
	if payload.Amount < 0 {
		jsonError(w, "amount must be non-negative", http.StatusBadRequest)
		return
	}

	// Perform a partial update
	updates := map[string]interface{}{
		"invoice_type": payload.InvoiceType,
		"issue_date":   payload.IssueDate,
		"due_date":     payload.DueDate,
		"status":       payload.Status,
		"amount":       payload.Amount,
		"currency":     payload.Currency,
		"customer_id":  payload.CustomerID,
		"vendor_id":    payload.VendorID,
		"updated_at":   time.Now(),
	}
	if err := h.DB.Model(&invoice).Updates(updates).Error; err != nil {
		jsonError(w, "Failed to update invoice", http.StatusInternalServerError)
		return
	}

	_ = utils.LogAction(h.DB, claims.TenantID, claims.UserID,
		"UPDATE_INVOICE", "Invoice", "Updated invoice")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}

// DeleteInvoice handles DELETE /invoices/{invoiceID}
func (h *InvoiceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "invoiceID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		jsonError(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)

	// Delete in one statement with tenant check
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		Delete(&models.Invoice{}).Error; err != nil {
		jsonError(w, "Failed to delete invoice", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DownloadInvoicePDF handles GET /invoices/{invoiceID}/download
func (h *InvoiceHandler) DownloadInvoicePDF(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "invoiceID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		jsonError(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ContextKeyClaims).(*auth.Claims)

	var invoice models.Invoice
	if err := h.DB.
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Invoice not found", http.StatusNotFound)
		} else {
			jsonError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	pdfBytes, err := utils.GenerateInvoicePDF(invoice)
	if err != nil {
		jsonError(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("invoice_%s.pdf", invoice.ID)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/pdf")
	_, _ = io.Copy(w, bytes.NewReader(pdfBytes))
}
