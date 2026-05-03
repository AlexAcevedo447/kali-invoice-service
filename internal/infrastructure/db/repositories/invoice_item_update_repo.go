package repositories

import (
	"context"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceItemUpdateRepo struct {
	db *bun.DB
}

func NewInvoiceItemUpdateRepo(db *bun.DB) ports.InvoiceItemUpdater {
	return &invoiceItemUpdateRepo{db: db}
}

func (r *invoiceItemUpdateRepo) Update(item *invoiceitem.InvoiceItem) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	model := mappers.InvoiceItemToModel(item)
	_, err = tx.NewUpdate().
		Model(model).
		Column("quantity", "unit_price", "tax_total", "subtotal", "total").
		WherePK().
		Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.NewDelete().Table("invoice_item_taxes").Where("invoice_item_id = ?", item.ID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if len(model.Taxes) > 0 {
		if _, err := tx.NewInsert().Model(&model.Taxes).Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := recalculateInvoiceTotalsTx(ctx, tx, item.InvoiceID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}