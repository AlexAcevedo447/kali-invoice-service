package repositories

import (
	"context"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceListRepo struct {
	db *bun.DB
}

func NewInvoiceListRepo(db *bun.DB) ports.InvoiceListReader {
	return &invoiceListRepo{db: db}
}

func (r *invoiceListRepo) GetAll(tenantID string, p ports.Pagination) ([]*invoice.Invoice, error) {
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.PageSize

	var ms []models.InvoiceModel
	err := r.db.NewSelect().
		Model(&ms).
		Relation("Items").
		Relation("Items.Taxes", orderItemTaxes).
		Relation("Taxes").
		Relation("Customer").
		Relation("Items.Product").
		Where("i.tenant_id = ?", tenantID).
		// Orden estable: sin ORDER BY, Postgres puede repetir u omitir filas entre páginas.
		OrderExpr("i.created_at DESC, i.id DESC").
		Limit(p.PageSize).
		Offset(offset).
		Scan(context.Background())
	if err != nil {
		return nil, err
	}

	result := make([]*invoice.Invoice, 0, len(ms))
	for i := range ms {
		result = append(result, mappers.ModelToInvoice(&ms[i]))
	}
	return result, nil
}

func (r *invoiceListRepo) Count(tenantID string) (int, error) {
	return r.db.NewSelect().
		Model((*models.InvoiceModel)(nil)).
		Where("tenant_id = ?", tenantID).
		Count(context.Background())
}
