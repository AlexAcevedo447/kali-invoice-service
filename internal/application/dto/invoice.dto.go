package dto

import (
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
)

type AppCreateInvoiceDTO struct {
	CustomerId string
	Items []entities.NewInvoiceItemDTO
	IssueDate *time.Time
}

type AppGetByIdInvoiceDTO struct {
	Id string
}