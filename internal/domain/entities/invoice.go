package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/dto/"
)


type InvoiceStatus string

const (
	StatusPending InvoiceStatus = "PENDING"
	StatusPaid InvoiceStatus = "PAID"
	StatusCanceled InvoiceStatus = "CANCELED"
)

type Invoice struct {
	Id string
	CustomerId string
	IssueDate time.Time
	DueDate time.Time
	Items []InvoiceItem
	Total float64
	Status InvoiceStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewInvoice(dto NewInvoiceDTO) (*Invoice, error) {

}