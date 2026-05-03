package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceStatusRepo struct {
	db *bun.DB
}

func NewInvoiceStatusRepo(db *bun.DB) ports.InvoiceStatusUpdater {
	return &invoiceStatusRepo{db: db}
}

func (r *invoiceStatusRepo) UpdateStatus(id string, apply func(*invoice.Invoice) error) (*invoice.Invoice, error) {
	ctx := context.Background()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	m := new(models.InvoiceModel)
	err = tx.NewSelect().
		Model(m).
		Where("i.id = ?", id).
		For("UPDATE").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, invoice.ErrNotFound
		}
		return nil, err
	}

	inv := mappers.ModelToInvoice(m)
	if err := apply(inv); err != nil {
		return nil, err
	}

	updated := mappers.InvoiceToModel(inv)
	if _, err := tx.NewUpdate().Model(updated).WherePK().Exec(ctx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return inv, nil
}
