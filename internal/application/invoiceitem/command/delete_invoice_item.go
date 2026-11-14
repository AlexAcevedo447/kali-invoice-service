package command

import "github.com/AlexAcevedo447/kali-invoice-service/internal/ports"

type DeleteInvoiceItemCommand struct {
	deleter ports.InvoiceItemDeleter
}

func NewDeleteInvoiceItemCommand(deleter ports.InvoiceItemDeleter) *DeleteInvoiceItemCommand {
	return &DeleteInvoiceItemCommand{deleter: deleter}
}

func (c *DeleteInvoiceItemCommand) Execute(tenantID, id string) error {
	return c.deleter.DeleteByID(tenantID, id)
}
