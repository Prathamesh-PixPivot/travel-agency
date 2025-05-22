package jobs

import (
	"log"
	"time"

	"travel-agency/internal/models"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// ReconcileInvoices recalculates payment totals and updates invoice statuses.
func ReconcileInvoices(db *gorm.DB) {
	var invoices []models.Invoice
	// Get invoices not marked as Paid or Canceled.
	if err := db.Where("status NOT IN ?", []string{"Paid", "Canceled"}).Find(&invoices).Error; err != nil {
		log.Printf("Error fetching invoices: %v", err)
		return
	}

	for _, invoice := range invoices {
		var totalPaid float64
		db.Model(&models.Payment{}).
			Where("invoice_id = ?", invoice.ID).
			Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalPaid)

		// Reconciliation logic:
		if totalPaid >= invoice.Amount {
			invoice.Status = "Paid"
		} else if time.Now().After(invoice.DueDate) {
			invoice.Status = "Overdue"
		} else {
			invoice.Status = "Outstanding"
		}

		if err := db.Save(&invoice).Error; err != nil {
			log.Printf("Failed to update invoice %d: %v", invoice.ID, err)
		}
	}
}

func StartCronJobs(db *gorm.DB) {
	c := cron.New()
	// Schedule the reconciliation job to run every hour.
	c.AddFunc("@hourly", func() { ReconcileInvoices(db) })
	c.Start()
}
