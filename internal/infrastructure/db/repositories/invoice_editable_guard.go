package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/uptrace/bun"
)

// ensureInvoiceEditableTx bloquea la factura dentro de la transacción y rechaza
// cualquier cambio de sus líneas si no está PENDING (PAID/CANCELED quedan cerradas)
// o si no existe.
func ensureInvoiceEditableTx(ctx context.Context, tx bun.Tx, invoiceID string) error {
	var status string
	err := tx.NewSelect().
		Table("invoices").
		Column("status").
		Where("id = ?", invoiceID).
		For("UPDATE").
		Scan(ctx, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return invoice.ErrNotFound
		}
		return err
	}
	inv := invoice.Invoice{Status: invoice.Status(status)}
	return inv.EnsureEditable()
}
