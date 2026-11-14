package invoice

import (
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
	ID string
	// TenantID es la empresa dueña de la factura; proviene siempre del token validado.
	TenantID   string
	CustomerID string
	// CustomerName y CustomerDocument son datos legibles del cliente (solo lectura,
	// se completan al consultar desde el catálogo de clientes del tenant).
	CustomerName     string
	CustomerDocument string
	IssueDate        time.Time
	DueDate          time.Time
	Items            []invoiceitem.InvoiceItem
	Subtotal         float64
	TaxTotal         float64
	Total            float64
	Status           Status
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ItemInput contiene los datos necesarios para crear un ítem al construir la factura.
type ItemInput struct {
	ItemID    string
	Quantity  float64
	UnitPrice float64
	Taxes     []invoiceitem.TaxInput
}

// New construye una factura válida desde cero.
func New(tenantID, customerID string, issueDate, dueDate time.Time, inputs []ItemInput) (*Invoice, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	if customerID == "" {
		return nil, ErrCustomerIDRequired
	}
	if len(inputs) == 0 {
		return nil, ErrItemsRequired
	}
	// Misma regla que al editar (UpdateHeader): el vencimiento no puede ser
	// anterior a la emisión.
	if dueDate.Before(issueDate) {
		return nil, ErrDueDateBeforeIssueDate
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
		TenantID:   tenantID,
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
// Solo las facturas PENDING son editables: PAID y CANCELED quedan cerradas.
func (i *Invoice) UpdateHeader(customerID string, dueDate time.Time) error {
	if err := i.EnsureEditable(); err != nil {
		return err
	}
	if customerID == "" {
		return ErrCustomerIDRequired
	}
	if dueDate.Before(i.IssueDate) {
		return ErrDueDateBeforeIssueDate
	}
	i.CustomerID = customerID
	i.DueDate = dueDate
	i.UpdatedAt = time.Now()
	return nil
}

// EnsureEditable devuelve error si la factura ya no admite cambios (PAID o CANCELED).
func (i *Invoice) EnsureEditable() error {
	switch i.Status {
	case StatusPaid:
		return ErrAlreadyPaid
	case StatusCanceled:
		return ErrAlreadyCanceled
	}
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
