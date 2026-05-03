package ports

import (
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
)

// InvoiceStatusChangedEvent representa el evento de negocio emitido al cambiar estado de factura.
type InvoiceStatusChangedEvent struct {
	InvoiceID       string           `json:"invoice_id"`
	PreviousStatus  invoice.Status   `json:"previous_status"`
	NewStatus       invoice.Status   `json:"new_status"`
	ChangedAt       time.Time        `json:"changed_at"`
	InvoiceSnapshot *invoice.Invoice `json:"invoice_snapshot"`
}

// InvoiceStatusEventPublisher publica eventos de cambio de estado a infraestructura de mensajeria.
type InvoiceStatusEventPublisher interface {
	PublishInvoiceStatusChanged(event InvoiceStatusChangedEvent) error
}
