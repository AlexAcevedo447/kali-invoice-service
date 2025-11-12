package entities

type NewItemDTO struct {
	Name string
	Quantity int
	Description string
	TaxRate float64
	UnitPrice float64
}