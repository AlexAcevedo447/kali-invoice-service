package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceItemByIDRepo struct {
	db *bun.DB
}

func NewInvoiceItemByIDRepo(db *bun.DB) ports.InvoiceItemByIDReader {
	return &invoiceItemByIDRepo{db: db}
}

func (r *invoiceItemByIDRepo) GetByID(tenantID, id string) (*invoiceitem.InvoiceItem, error) {
	m := new(models.InvoiceItemModel)
	err := r.db.NewSelect().
		Model(m).
		Relation("Taxes").
		Relation("Product").
		// La línea pertenece al tenant de su factura.
		Join("JOIN invoices AS inv ON inv.id = ii.invoice_id").
		Where("ii.id = ?", id).
		Where("inv.tenant_id = ?", tenantID).
		Scan(context.Background())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, invoiceitem.ErrNotFound
		}
		return nil, err
	}
	return mappers.ModelToInvoiceItem(m), nil
}
