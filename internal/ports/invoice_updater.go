package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"

// InvoiceUpdater persiste actualizaciones de encabezado de facturas.
type InvoiceUpdater interface {
	Update(inv *invoice.Invoice) error
}
