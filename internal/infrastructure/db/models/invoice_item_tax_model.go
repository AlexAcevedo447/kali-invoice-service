package models

import "github.com/uptrace/bun"

type InvoiceItemTaxModel struct {
	bun.BaseModel `bun:"table:invoice_item_taxes,alias:iit"`

	ID            string  `bun:"id,pk,type:uuid"`
	InvoiceID     *string `bun:"invoice_id,type:uuid,nullzero"`
	InvoiceItemID *string `bun:"invoice_item_id,type:uuid,nullzero"`
	Code          string  `bun:"code,notnull"`
	Kind          string  `bun:"kind,notnull"`
	Rate          float64 `bun:"rate,notnull"`
	Amount        float64 `bun:"amount,notnull"`
	// Position conserva el orden en que se registraron los impuestos de la línea
	// (CA-13.2). Las filas anteriores a la migración v9 quedan en 0.
	Position int `bun:"position,notnull,default:0"`
}
