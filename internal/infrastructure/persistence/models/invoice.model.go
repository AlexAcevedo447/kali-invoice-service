package models

import "time"

type InvoiceModel struct {
	Id string
	CustomerId string
	IssueDate time.Time
	DueDate time.Time
	Items []InvoiceItemModel
	Total float64
	Status string
	CreatedAt time.Time
	UpdatedAt time.Time
}