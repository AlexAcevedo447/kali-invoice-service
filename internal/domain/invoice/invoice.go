package invoice

import (
	"errors"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/google/uuid"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusPaid     Status = "PAID"
	StatusCanceled Status = "CANCELED"
)

type Invoice struct {
	ID         string
	CustomerID string
	IssueDate  time.Time
	DueDate    time.Time
	Items      []invoiceitem.InvoiceItem
	Subtotal   float64
	TaxTotal   float64
	Total      float64
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ItemInput contiene los datos necesarios para crear un ítem al construir la factura.
type ItemInput struct {
	ItemID    string
	Quantity  float64
	UnitPrice float64
	Taxes     []invoiceitem.TaxInput
}

// New construye una factura válida desde cero.
func New(customerID string, issueDate, dueDate time.Time, inputs []ItemInput) (*Invoice, error) {
	if customerID == "" {
		return nil, errors.New("customer_id is required")
	}
	if len(inputs) == 0 {
		return nil, errors.New("invoice must have at least one item")
	}

	invoiceID := uuid.New().String()
	now := time.Now()

	items := make([]invoiceitem.InvoiceItem, 0, len(inputs))
	for _, in := range inputs {
		item, err := invoiceitem.New(invoiceID, in.ItemID, in.Quantity, in.UnitPrice, in.Taxes)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	inv := &Invoice{
		ID:         invoiceID,
		CustomerID: customerID,
		IssueDate:  issueDate,
		DueDate:    dueDate,
		Items:      items,
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	inv.recalculateTotal()
	return inv, nil
}

func (i *Invoice) recalculateTotal() {
	var subtotal float64
	var taxTotal float64
	for _, item := range i.Items {
		subtotal += item.Subtotal
		taxTotal += item.TaxTotal
	}
	i.Subtotal = subtotal
	i.TaxTotal = taxTotal
	i.Total = subtotal + taxTotal
}

// UpdateHeader actualiza los datos principales de la factura.
func (i *Invoice) UpdateHeader(customerID string, dueDate time.Time) error {
	if customerID == "" {
		return errors.New("customer_id is required")
	}
	if dueDate.Before(i.IssueDate) {
		return errors.New("due_date cannot be before issue_date")
	}
	i.CustomerID = customerID
	i.DueDate = dueDate
	i.UpdatedAt = time.Now()
	return nil
}

// MarkAsPaid transiciona la factura al estado PAID.
func (i *Invoice) MarkAsPaid() error {
	if i.Status == StatusCanceled {
		return ErrAlreadyCanceled
	}
	if i.Status == StatusPaid {
		return ErrAlreadyPaid
	}
	i.Status = StatusPaid
	i.UpdatedAt = time.Now()
	return nil
}

// Cancel transiciona la factura al estado CANCELED.
func (i *Invoice) Cancel() error {
	if i.Status == StatusPaid {
		return ErrAlreadyPaid
	}
	if i.Status == StatusCanceled {
		return ErrAlreadyCanceled
	}
	i.Status = StatusCanceled
	i.UpdatedAt = time.Now()
	return nil
}
