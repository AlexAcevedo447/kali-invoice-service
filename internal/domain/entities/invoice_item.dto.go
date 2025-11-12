package entities

type NewInvoiceItemDTO struct {
	Item *Item
	Quantity int
	UnitPrice float64
}