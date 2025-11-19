package entities

type NewInvoiceItemDTO struct {
	InvoiceId string
	Item *Item
	Quantity float64
	UnitPrice float64
}