package command

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type CreateInvoiceItemInput struct {
	InvoiceID string                `json:"invoice_id"`
	ItemID    string                `json:"item_id"`
	Quantity  float64               `json:"quantity"`
	UnitPrice float64               `json:"unit_price"`
	Taxes     []InvoiceItemTaxInput `json:"taxes,omitempty"`
}

type InvoiceItemTaxInput struct {
	Code string  `json:"code"`
	Kind string  `json:"kind,omitempty"`
	Rate float64 `json:"rate"`
}

type CreateInvoiceItemCommand struct {
	repo     ports.InvoiceItemCreator
	products ports.ProductRepository
}

func NewCreateInvoiceItemCommand(repo ports.InvoiceItemCreator, products ports.ProductRepository) *CreateInvoiceItemCommand {
	return &CreateInvoiceItemCommand{repo: repo, products: products}
}

func (c *CreateInvoiceItemCommand) Execute(tenantID string, input CreateInvoiceItemInput) (*invoiceitem.InvoiceItem, error) {
	// El producto debe pertenecer al tenant; la factura se valida en el repositorio.
	missing, err := c.products.MissingIDs(tenantID, []string{input.ItemID})
	if err != nil {
		return nil, err
	}
	if len(missing) > 0 {
		return nil, catalog.ErrProductNotFound
	}

	taxes := mapTaxInputs(input.Taxes)
	item, err := invoiceitem.New(input.InvoiceID, input.ItemID, input.Quantity, input.UnitPrice, taxes)
	if err != nil {
		return nil, err
	}
	if err := c.repo.Save(tenantID, item); err != nil {
		return nil, err
	}
	return item, nil
}
