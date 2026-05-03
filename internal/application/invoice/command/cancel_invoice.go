package command

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

// CancelInvoiceCommand transiciona una factura al estado CANCELED.
// Es la única forma de "eliminar" una factura en el sistema.
type CancelInvoiceCommand struct {
	commandRepo ports.InvoiceStatusUpdater
	publisher   ports.InvoiceStatusEventPublisher
}

func NewCancelInvoiceCommand(c ports.InvoiceStatusUpdater, publisher ports.InvoiceStatusEventPublisher) *CancelInvoiceCommand {
	return &CancelInvoiceCommand{commandRepo: c, publisher: publisher}
}

func (cmd *CancelInvoiceCommand) Execute(id string) error {
	var previousStatus invoice.Status
	updated, err := cmd.commandRepo.UpdateStatus(id, func(inv *invoice.Invoice) error {
		previousStatus = inv.Status
		return inv.Cancel()
	})
	if err != nil {
		return err
	}

	return cmd.publisher.PublishInvoiceStatusChanged(ports.InvoiceStatusChangedEvent{
		InvoiceID:       updated.ID,
		PreviousStatus:  previousStatus,
		NewStatus:       updated.Status,
		ChangedAt:       updated.UpdatedAt,
		InvoiceSnapshot: updated,
	})
}
