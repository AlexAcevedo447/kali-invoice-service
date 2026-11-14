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
	reader    ports.InvoiceByIDReader
	updater   ports.InvoiceUpdater
	customers ports.CustomerRepository
}

func NewUpdateInvoiceCommand(reader ports.InvoiceByIDReader, updater ports.InvoiceUpdater, customers ports.CustomerRepository) *UpdateInvoiceCommand {
	return &UpdateInvoiceCommand{reader: reader, updater: updater, customers: customers}
}

func (c *UpdateInvoiceCommand) Execute(tenantID, id string, input UpdateInvoiceInput) (*invoice.Invoice, error) {
	inv, err := c.reader.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	customerID := inv.CustomerID
	if input.CustomerID != nil && *input.CustomerID != inv.CustomerID {
		customerID = *input.CustomerID
		if customerID != "" {
			if err := ensureCustomerInTenant(c.customers, tenantID, customerID); err != nil {
				return nil, err
			}
		}
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

	return c.reader.GetByID(tenantID, id)
}
