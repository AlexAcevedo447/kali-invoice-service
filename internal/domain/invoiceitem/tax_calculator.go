package invoiceitem

import (
	"github.com/google/uuid"
)

func calculateTaxes(invoiceID, invoiceItemID string, subtotal float64, taxes []TaxInput) ([]Tax, float64, error) {
	result := make([]Tax, 0, len(taxes))
	var taxTotal float64

	for _, t := range taxes {
		if t.Code == "" {
			return nil, 0, ErrTaxCodeRequired
		}
		if t.Rate < 0 {
			return nil, 0, ErrTaxRateInvalid
		}
		kind, err := normalizeTaxKind(t.Kind)
		if err != nil {
			return nil, 0, err
		}
		amount := subtotal * t.Rate
		tax := Tax{
			ID:            uuid.New().String(),
			InvoiceID:     invoiceID,
			InvoiceItemID: invoiceItemID,
			Code:          t.Code,
			Kind:          kind,
			Rate:          t.Rate,
			Amount:        amount,
		}
		result = append(result, tax)
		taxTotal += signedAmount(kind, amount)
	}

	return result, taxTotal, nil
}

func normalizeTaxKind(kind string) (TaxKind, error) {
	switch TaxKind(kind) {
	case "", TaxKindDebit:
		return TaxKindDebit, nil
	case TaxKindCredit:
		return TaxKindCredit, nil
	default:
		return "", ErrTaxKindInvalid
	}
}

func signedAmount(kind TaxKind, amount float64) float64 {
	if kind == TaxKindCredit {
		return -amount
	}
	return amount
}
