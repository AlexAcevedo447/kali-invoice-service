package command

import (
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type CreateInvoiceInput struct {
	// TenantID proviene del token validado, nunca del cuerpo de la petición.
	TenantID   string
	CustomerID string
	IssueDate  *time.Time
	DueDate    *time.Time
	Items      []InvoiceItemInput
}

type InvoiceItemInput struct {
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

type CreateInvoiceCommand struct {
	repo      ports.InvoiceCreator
	reader    ports.InvoiceByIDReader
	customers ports.CustomerRepository
	products  ports.ProductRepository
}

func NewCreateInvoiceCommand(
	repo ports.InvoiceCreator,
	reader ports.InvoiceByIDReader,
	customers ports.CustomerRepository,
	products ports.ProductRepository,
) *CreateInvoiceCommand {
	return &CreateInvoiceCommand{repo: repo, reader: reader, customers: customers, products: products}
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

	// Cliente y productos deben pertenecer al mismo tenant de la factura.
	if err := ensureCustomerInTenant(c.customers, input.TenantID, input.CustomerID); err != nil {
		return nil, err
	}
	productIDs := make([]string, 0, len(input.Items))
	itemInputs := make([]invoice.ItemInput, 0, len(input.Items))
	for _, it := range input.Items {
		productIDs = append(productIDs, it.ItemID)
		itemInputs = append(itemInputs, invoice.ItemInput{
			ItemID:    it.ItemID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Taxes:     mapInvoiceItemTaxes(it.Taxes),
		})
	}
	if err := ensureProductsInTenant(c.products, input.TenantID, productIDs); err != nil {
		return nil, err
	}

	inv, err := invoice.New(input.TenantID, input.CustomerID, issueDate, dueDate, itemInputs)
	if err != nil {
		return nil, err
	}

	if err := c.repo.Save(inv); err != nil {
		return nil, err
	}

	// Se relee para devolver también los datos legibles de cliente y productos.
	return c.reader.GetByID(input.TenantID, inv.ID)
}
