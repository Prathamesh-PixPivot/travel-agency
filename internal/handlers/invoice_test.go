package handlers

import (
	"bytes"
	"context"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"travel-agency/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type contextKey string

const claimsKey contextKey = "claims"

// SetupTestDB creates an in-memory SQLite DB for testing.
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	// Auto-migrate relevant models.
	err = db.AutoMigrate(&models.Invoice{}, &models.Payment{}, &models.AuditLog{})
	assert.NoError(t, err)
	return db
}

func TestUpdateInvoice(t *testing.T) {
	// Setup a test database.
	db := SetupTestDB(t)

	// Seed the database with a sample invoice.
	invoice := models.Invoice{
		TenantID:    1,
		InvoiceType: "sale",
		IssueDate:   time.Now().Add(-24 * time.Hour),
		DueDate:     time.Now().Add(24 * time.Hour),
		Status:      "Draft",
		Amount:      1000,
		Currency:    "USD",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}
	assert.NoError(t, db.Create(&invoice).Error)

	// Create an invoice handler.
	handler := NewInvoiceHandler(db)

	// Create a new HTTP request that updates the invoice.
	updatePayload := map[string]interface{}{
		"invoice_type": "sale",
		"issue_date":   time.Now().Format(time.RFC3339),
		"due_date":     time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		"status":       "Sent",
		"amount":       1100,
		"currency":     "USD",
		"customer_id":  invoice.CustomerID,
		"vendor_id":    invoice.VendorID,
	}
	payloadBytes, _ := json.Marshal(updatePayload)
	req := httptest.NewRequest(http.MethodPut, "/api/invoices/1", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// We simulate authentication by manually setting tenant information in the context.
	// In a real test, you would use your AuthMiddleware; here we insert a dummy claim.
	ctx := req.Context()
	// For testing, assume claims are of type *auth.Claims; create a dummy.
	dummyClaims := &models.DummyClaims{
		TenantID: 1,
		UserID:   42,
	}
	// Instead of models.DummyClaims, if you have defined your own auth.Claims, use it.
	ctx = contextWithDummyClaims(ctx, dummyClaims)
	req = req.WithContext(ctx)

	// Create a ResponseRecorder to record the response.
	rr := httptest.NewRecorder()

	// Create a Chi router and mount the handler.
	r := chi.NewRouter()
	r.Put("/api/invoices/{invoiceID}", handler.UpdateInvoice)
	r.ServeHTTP(rr, req)

	// Validate the response.
	assert.Equal(t, http.StatusOK, rr.Code)
	var updatedInvoice models.Invoice
	err := json.Unmarshal(rr.Body.Bytes(), &updatedInvoice)
	assert.NoError(t, err)
	assert.Equal(t, "Sent", updatedInvoice.Status)
	assert.Equal(t, float64(1100), updatedInvoice.Amount)

	// Optionally, check that audit log was created.
	var audit models.AuditLog
	err = db.First(&audit, "tenant_id = ? AND user_id = ?", 1, 42).Error
	assert.NoError(t, err)
}

// contextWithDummyClaims is a helper to add dummy claims into a request context for testing.
func contextWithDummyClaims(ctx context.Context, claims interface{}) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}
