package command

import (
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type CreateInvoiceInput struct {
	CustomerID string
	IssueDate  *time.Time
	DueDate    *time.Time
	Items      []InvoiceItemInput
}

type InvoiceItemInput struct {
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

type CreateInvoiceCommand struct {
	repo ports.InvoiceCreator
}

func NewCreateInvoiceCommand(repo ports.InvoiceCreator) *CreateInvoiceCommand {
	return &CreateInvoiceCommand{repo: repo}
}

func (c *CreateInvoiceCommand) Execute(input CreateInvoiceInput) (*invoice.Invoice, error) {
	issueDate := time.Now()
	if input.IssueDate != nil {
		issueDate = *input.IssueDate
	}
	dueDate := issueDate.AddDate(0, 0, 30)
	if input.DueDate != nil {
		dueDate = *input.DueDate
	}

	itemInputs := make([]invoice.ItemInput, 0, len(input.Items))
	for _, it := range input.Items {
		taxes := make([]invoiceitem.TaxInput, 0, len(it.Taxes))
		for _, tax := range it.Taxes {
			taxes = append(taxes, invoiceitem.TaxInput{Code: tax.Code, Kind: tax.Kind, Rate: tax.Rate})
		}
		itemInputs = append(itemInputs, invoice.ItemInput{
			ItemID:    it.ItemID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Taxes:     taxes,
		})
	}

	inv, err := invoice.New(input.CustomerID, issueDate, dueDate, itemInputs)
	if err != nil {
		return nil, err
	}

	if err := c.repo.Save(inv); err != nil {
		return nil, err
	}

	return inv, nil
}
