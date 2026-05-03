package command

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type CreateInvoiceItemInput struct {
	InvoiceID string  `json:"invoice_id"`
	ItemID    string  `json:"item_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Taxes     []InvoiceItemTaxInput `json:"taxes,omitempty"`
}

type InvoiceItemTaxInput struct {
	Code string  `json:"code"`
	Kind string  `json:"kind,omitempty"`
	Rate float64 `json:"rate"`
}

type CreateInvoiceItemCommand struct {
	repo ports.InvoiceItemCreator
}

func NewCreateInvoiceItemCommand(repo ports.InvoiceItemCreator) *CreateInvoiceItemCommand {
	return &CreateInvoiceItemCommand{repo: repo}
}

func (c *CreateInvoiceItemCommand) Execute(input CreateInvoiceItemInput) (*invoiceitem.InvoiceItem, error) {
	taxes := make([]invoiceitem.TaxInput, 0, len(input.Taxes))
	for _, t := range input.Taxes {
		taxes = append(taxes, invoiceitem.TaxInput{Code: t.Code, Kind: t.Kind, Rate: t.Rate})
	}
	item, err := invoiceitem.New(input.InvoiceID, input.ItemID, input.Quantity, input.UnitPrice, taxes)
	if err != nil {
		return nil, err
	}
	if err := c.repo.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}