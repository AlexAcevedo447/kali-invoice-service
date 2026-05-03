package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"

// InvoiceStatusUpdater aplica transiciones de estado de forma atómica.
type InvoiceStatusUpdater interface {
	// UpdateStatus carga la factura, aplica la transición de dominio y persiste,
	// todo dentro de una transacción con bloqueo FOR UPDATE para evitar condiciones de carrera.
	UpdateStatus(id string, apply func(*invoice.Invoice) error) (*invoice.Invoice, error)
}