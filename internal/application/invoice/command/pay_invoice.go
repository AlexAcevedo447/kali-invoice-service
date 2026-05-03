package command

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

// PayInvoiceCommand transiciona una factura al estado PAID.
type PayInvoiceCommand struct {
	commandRepo ports.InvoiceStatusUpdater
	publisher   ports.InvoiceStatusEventPublisher
}

func NewPayInvoiceCommand(c ports.InvoiceStatusUpdater, publisher ports.InvoiceStatusEventPublisher) *PayInvoiceCommand {
	return &PayInvoiceCommand{commandRepo: c, publisher: publisher}
}

func (cmd *PayInvoiceCommand) Execute(id string) error {
	var previousStatus invoice.Status
	updated, err := cmd.commandRepo.UpdateStatus(id, func(inv *invoice.Invoice) error {
		previousStatus = inv.Status
		return inv.MarkAsPaid()
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
