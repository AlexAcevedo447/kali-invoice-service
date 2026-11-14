package repositories

import (
	"context"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceUpdateRepo struct {
	db *bun.DB
}

func NewInvoiceUpdateRepo(db *bun.DB) ports.InvoiceUpdater {
	return &invoiceUpdateRepo{db: db}
}

func (r *invoiceUpdateRepo) Update(inv *invoice.Invoice) error {
	model := mappers.InvoiceToModel(inv)
	res, err := r.db.NewUpdate().
		Model(model).
		Column("customer_id", "issue_date", "due_date", "subtotal", "tax_total", "total", "status", "updated_at").
		WherePK().
		Where("tenant_id = ?", inv.TenantID).
		Exec(context.Background())
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return invoice.ErrNotFound
	}
	return nil
}
