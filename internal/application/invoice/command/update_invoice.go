package command

import (
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type UpdateInvoiceInput struct {
	CustomerID *string
	DueDate    *time.Time
}

type UpdateInvoiceCommand struct {
	reader  ports.InvoiceByIDReader
	updater ports.InvoiceUpdater
}

func NewUpdateInvoiceCommand(reader ports.InvoiceByIDReader, updater ports.InvoiceUpdater) *UpdateInvoiceCommand {
	return &UpdateInvoiceCommand{reader: reader, updater: updater}
}

func (c *UpdateInvoiceCommand) Execute(id string, input UpdateInvoiceInput) (*invoice.Invoice, error) {
	inv, err := c.reader.GetByID(id)
	if err != nil {
		return nil, err
	}

	customerID := inv.CustomerID
	if input.CustomerID != nil {
		customerID = *input.CustomerID
	}
	dueDate := inv.DueDate
	if input.DueDate != nil {
		dueDate = *input.DueDate
	}

	if err := inv.UpdateHeader(customerID, dueDate); err != nil {
		return nil, err
	}
	if err := c.updater.Update(inv); err != nil {
		return nil, err
	}
	return inv, nil
}
