package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)


type InvoiceStatus string

const (
	StatusPending InvoiceStatus = "PENDING"
	StatusPaid InvoiceStatus = "PAID"
	StatusCanceled InvoiceStatus = "CANCELED"
)

type Invoice struct {
	Id string `gorm:"primaryKey;type:uuid" json:"id"`
	CustomerId string `gorm:"type:uuid;not null" json:"customer_id"`
	IssueDate time.Time `gorm:"not null" json:"issue_date"`
	DueDate time.Time `gorm:"not null" json:"due_date"`
	Items []InvoiceItem `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"items"`
	Total float64 `gorm:"not null" json:"total"`
	Status InvoiceStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewInvoice(dto NewInvoiceDTO) (*Invoice, error) {
	if dto.CustomerID == "" {
		return nil, errors.New("CustomerId is required")
	}

	invoice := &Invoice{
		Id: uuid.New().String(),
		CustomerId: dto.CustomerID,
		IssueDate: dto.IssueDate,
		DueDate: dto.DueDate,
		Items: dto.Items,
		Status: StatusPending,
	}

	invoice.CalculateTotal()
	return invoice, nil
}

func (i *Invoice) CalculateTotal() (float64){
	var total float64
	
	for idx := range i.Items {
		i.Items[idx].CalculateSubtotal()
		total += i.Items[idx].Subtotal
	}

	i.Total = total

	return total
}

func (i *Invoice) MarkAsPaid(){
	i.Status = StatusPaid
}

func (i *Invoice) Cancel(){
	i.Status = StatusCanceled
}

func (i *Invoice) Validate() error{
	if i.CustomerId == ""{
		return errors.New("CustomerId is required")
	}

	if len(i.Items) ==0{
		return errors.New("Invoice must have at least one Item")
	}

	return nil
}