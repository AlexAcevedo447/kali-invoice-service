package models

import (
	"time"

	"github.com/uptrace/bun"
)

type InvoiceModel struct {
	bun.BaseModel `bun:"table:invoices,alias:i"`

	ID string `bun:"id,pk,type:uuid"`
	// TenantID es la empresa dueña de la factura. Las filas anteriores a la
	// migración 6 quedan con NULL y no son visibles para ningún tenant.
	TenantID   string                `bun:"tenant_id,type:text,nullzero"`
	CustomerID string                `bun:"customer_id,notnull,type:uuid"`
	IssueDate  time.Time             `bun:"issue_date,notnull"`
	DueDate    time.Time             `bun:"due_date,notnull"`
	Subtotal   float64               `bun:"subtotal,notnull"`
	TaxTotal   float64               `bun:"tax_total,notnull"`
	Total      float64               `bun:"total,notnull"`
	Status     string                `bun:"status,notnull"`
	CreatedAt  time.Time             `bun:"created_at,notnull"`
	UpdatedAt  time.Time             `bun:"updated_at,notnull"`
	Items      []InvoiceItemModel    `bun:"rel:has-many,join:id=invoice_id"`
	Taxes      []InvoiceItemTaxModel `bun:"rel:has-many,join:id=invoice_id"`
	// Customer solo se usa para leer datos legibles del cliente.
	Customer *CustomerModel `bun:"rel:belongs-to,join:customer_id=id"`
}
