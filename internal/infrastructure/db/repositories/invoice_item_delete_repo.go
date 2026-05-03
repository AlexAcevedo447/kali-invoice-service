package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceItemDeleteRepo struct {
	db *bun.DB
}

func NewInvoiceItemDeleteRepo(db *bun.DB) ports.InvoiceItemDeleter {
	return &invoiceItemDeleteRepo{db: db}
}

func (r *invoiceItemDeleteRepo) DeleteByID(id string) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var invoiceID string
	if err := tx.NewSelect().Table("invoice_items").Column("invoice_id").Where("id = ?", id).Scan(ctx, &invoiceID); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return invoiceitem.ErrNotFound
		}
		return err
	}

	if _, err := tx.NewDelete().Table("invoice_item_taxes").Where("invoice_item_id = ?", id).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	res, err := tx.NewDelete().Table("invoice_items").Where("id = ?", id).Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if affected == 0 {
		_ = tx.Rollback()
		return invoiceitem.ErrNotFound
	}

	if err := recalculateInvoiceTotalsTx(ctx, tx, invoiceID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}