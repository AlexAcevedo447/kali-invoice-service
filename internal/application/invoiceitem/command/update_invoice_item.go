package command

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type UpdateInvoiceItemInput struct {
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Taxes     []InvoiceItemTaxInput `json:"taxes,omitempty"`
}

type UpdateInvoiceItemCommand struct {
	reader  ports.InvoiceItemByIDReader
	updater ports.InvoiceItemUpdater
}

func NewUpdateInvoiceItemCommand(reader ports.InvoiceItemByIDReader, updater ports.InvoiceItemUpdater) *UpdateInvoiceItemCommand {
	return &UpdateInvoiceItemCommand{reader: reader, updater: updater}
}

func (c *UpdateInvoiceItemCommand) Execute(id string, input UpdateInvoiceItemInput) (*invoiceitem.InvoiceItem, error) {
	item, err := c.reader.GetByID(id)
	if err != nil {
		return nil, err
	}
	taxes := mapTaxInputs(input.Taxes)
	if err := item.Update(input.Quantity, input.UnitPrice, taxes); err != nil {
		return nil, err
	}
	if err := c.updater.Update(item); err != nil {
		return nil, err
	}
	return item, nil
}