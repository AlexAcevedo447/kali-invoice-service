package repositories

import (
	"context"

	"github.com/uptrace/bun"
)

type itemTotalsRow struct {
	Subtotal     float64 `bun:"subtotal"`
	ItemTaxTotal float64 `bun:"item_tax_total"`
}

type headerTaxTotalsRow struct {
	HeaderTaxTotal float64 `bun:"header_tax_total"`
}

func recalculateInvoiceTotalsTx(ctx context.Context, tx bun.Tx, invoiceID string) error {
	itemTotals := new(itemTotalsRow)
	err := tx.NewSelect().
		TableExpr("invoice_items AS ii").
		ColumnExpr("COALESCE(SUM(ii.subtotal), 0) AS subtotal").
		ColumnExpr("COALESCE(SUM(ii.tax_total), 0) AS item_tax_total").
		Where("ii.invoice_id = ?", invoiceID).
		Scan(ctx, itemTotals)
	if err != nil {
		return err
	}

	headerTaxes := new(headerTaxTotalsRow)
	err = tx.NewSelect().
		TableExpr("invoice_item_taxes AS iit").
		ColumnExpr("COALESCE(SUM(CASE WHEN iit.kind = 'CREDIT' THEN -iit.amount ELSE iit.amount END), 0) AS header_tax_total").
		Where("iit.invoice_id = ?", invoiceID).
		Where("iit.invoice_item_id IS NULL").
		Scan(ctx, headerTaxes)
	if err != nil {
		return err
	}

	taxTotal := itemTotals.ItemTaxTotal + headerTaxes.HeaderTaxTotal
	total := itemTotals.Subtotal + taxTotal

	_, err = tx.NewUpdate().
		Table("invoices").
		Set("subtotal = ?", itemTotals.Subtotal).
		Set("tax_total = ?", taxTotal).
		Set("total = ?", total).
		Set("updated_at = NOW()").
		Where("id = ?", invoiceID).
		Exec(ctx)
	return err
}
