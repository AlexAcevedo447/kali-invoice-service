package repositories

import (
	"context"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceCreateRepo struct {
	db *bun.DB
}

func NewInvoiceCreateRepo(db *bun.DB) ports.InvoiceCreator {
	return &invoiceCreateRepo{db: db}
}

func (r *invoiceCreateRepo) Save(inv *invoice.Invoice) error {
	ctx := context.Background()
	model := mappers.InvoiceToModel(inv)

	if _, err := r.db.NewInsert().Model(model).Exec(ctx); err != nil {
		return err
	}
	if len(model.Items) > 0 {
		if _, err := r.db.NewInsert().Model(&model.Items).Exec(ctx); err != nil {
			return err
		}

		allTaxes := make([]models.InvoiceItemTaxModel, 0)
		for _, item := range model.Items {
			for _, tax := range item.Taxes {
				allTaxes = append(allTaxes, tax)
			}
		}
		if len(allTaxes) > 0 {
			if _, err := r.db.NewInsert().Model(&allTaxes).Exec(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}