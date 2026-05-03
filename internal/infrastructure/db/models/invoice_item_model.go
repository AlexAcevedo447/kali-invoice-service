package models

import "github.com/uptrace/bun"

type InvoiceItemModel struct {
	bun.BaseModel `bun:"table:invoice_items,alias:ii"`

	ID        string  `bun:"id,pk,type:uuid"`
	InvoiceID string  `bun:"invoice_id,notnull,type:uuid"`
	ItemID    string  `bun:"item_id,notnull,type:uuid"`
	Quantity  float64 `bun:"quantity,notnull"`
	UnitPrice float64 `bun:"unit_price,notnull"`
	TaxTotal  float64 `bun:"tax_total,notnull"`
	Subtotal  float64 `bun:"subtotal,notnull"`
	Total     float64 `bun:"total,notnull"`
	Taxes     []InvoiceItemTaxModel `bun:"rel:has-many,join:id=invoice_item_id"`
}
