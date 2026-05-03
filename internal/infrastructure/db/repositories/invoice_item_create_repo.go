package repositories

import (
	"context"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceItemCreateRepo struct {
	db *bun.DB
}

func NewInvoiceItemCreateRepo(db *bun.DB) ports.InvoiceItemCreator {
	return &invoiceItemCreateRepo{db: db}
}

func (r *invoiceItemCreateRepo) Save(item *invoiceitem.InvoiceItem) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	model := mappers.InvoiceItemToModel(item)
	if _, err := tx.NewInsert().Model(model).Exec(ctx); err != nil {
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