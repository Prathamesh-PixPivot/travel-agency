package utils

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"travel-agency/internal/models"
)

// GenerateInvoicePDF builds a PDF document for the given invoice and returns the PDF as a byte slice.
func GenerateInvoicePDF(invoice models.Invoice) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Invoice")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice ID: %d", invoice.ID))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Type: %s", invoice.InvoiceType))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Issue Date: %s", invoice.IssueDate.Format("2006-01-02")))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Due Date: %s", invoice.DueDate.Format("2006-01-02")))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Amount: %.2f %s", invoice.Amount, invoice.Currency))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Status: %s", invoice.Status))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
