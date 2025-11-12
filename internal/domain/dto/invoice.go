package dto

import "time"

type NewInvoiceDTO struct {
	CustomerID string
	IssueDate  time.Time
	DueDate    time.Time
	Items      []InvoiceItem
}
