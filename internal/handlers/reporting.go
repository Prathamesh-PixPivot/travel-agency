// internal/handlers/reporting.go
package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type ReportingHandler struct {
	DB *gorm.DB
}

func NewReportingHandler(db *gorm.DB) *ReportingHandler {
	return &ReportingHandler{DB: db}
}

// SalesReport represents aggregated sales data.
type SalesReport struct {
	TotalRevenue float64 `json:"total_revenue"`
	TotalTrips   int     `json:"total_trips"`
}

// GetSalesReport returns a simple aggregated report.
func (h *ReportingHandler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	var report SalesReport
	// Example raw query: adjust per your schema.
	row := h.DB.Raw(`SELECT COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_trips FROM invoices WHERE status = 'Paid' AND tenant_id = ?`,
		r.Context().Value("tenantID")).Row()
	if err := row.Scan(&report.TotalRevenue, &report.TotalTrips); err != nil {
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
