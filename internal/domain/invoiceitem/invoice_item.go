package invoiceitem

import (
	"github.com/google/uuid"
)

type InvoiceItem struct {
	ID        string
	InvoiceID string
	ItemID    string
	// ProductCode y ProductName son datos legibles del producto (solo lectura).
	ProductCode string
	ProductName string
	Quantity    float64
	UnitPrice   float64
	Taxes       []Tax
	TaxTotal    float64
	Subtotal    float64
	Total       float64
}

func New(invoiceID, itemID string, quantity, unitPrice float64, taxes []TaxInput) (*InvoiceItem, error) {
	if invoiceID == "" {
		return nil, ErrInvoiceIDRequired
	}
	if itemID == "" {
		return nil, ErrItemIDRequired
	}
	if quantity <= 0 {
		return nil, ErrQuantityInvalid
	}
	if unitPrice <= 0 {
		return nil, ErrUnitPriceInvalid
	}

	subtotal := quantity * unitPrice
	calculatedTaxes, taxTotal, err := calculateTaxes(invoiceID, "", subtotal, taxes)
	if err != nil {
		return nil, err
	}

	itemIDGenerated := uuid.New().String()
	for idx := range calculatedTaxes {
		calculatedTaxes[idx].InvoiceItemID = itemIDGenerated
	}

	return &InvoiceItem{
		ID:        itemIDGenerated,
		InvoiceID: invoiceID,
		ItemID:    itemID,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		Taxes:     calculatedTaxes,
		TaxTotal:  taxTotal,
		Subtotal:  subtotal,
		Total:     subtotal + taxTotal,
	}, nil
}

func (i *InvoiceItem) Update(quantity, unitPrice float64, taxes []TaxInput) error {
	if quantity <= 0 {
		return ErrQuantityInvalid
	}
	if unitPrice <= 0 {
		return ErrUnitPriceInvalid
	}
	subtotal := quantity * unitPrice
	calculatedTaxes, taxTotal, err := calculateTaxes(i.InvoiceID, i.ID, subtotal, taxes)
	if err != nil {
		return err
	}
	i.Quantity = quantity
	i.UnitPrice = unitPrice
	i.Taxes = calculatedTaxes
	i.TaxTotal = taxTotal
	i.Subtotal = subtotal
	i.Total = subtotal + taxTotal
	return nil
}
