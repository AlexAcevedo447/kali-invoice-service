package invoiceitem

import (
	"github.com/google/uuid"
	"strings"
)

func calculateTaxes(invoiceID, invoiceItemID string, subtotal float64, taxes []TaxInput) ([]Tax, float64, error) {
	result := make([]Tax, 0, len(taxes))
	var taxTotal float64
	seen := make(map[string]bool, len(taxes))

	for _, t := range taxes {
		// D4-03: código de texto libre, obligatorio y normalizado (sin espacios
		// exteriores, en mayúsculas). D4-02: no se repite en la misma línea.
		code := NormalizeTaxCode(t.Code)
		if code == "" {
			return nil, 0, ErrTaxCodeRequired
		}
		if seen[code] {
			return nil, 0, ErrTaxCodeDuplicated
		}
		seen[code] = true
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
			Code:          code,
			Kind:          kind,
			Rate:          t.Rate,
			Amount:        amount,
		}
		result = append(result, tax)
		taxTotal += signedAmount(kind, amount)
	}

	return result, taxTotal, nil
}

// NormalizeTaxCode: sin espacios exteriores y en mayúsculas.
func NormalizeTaxCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// normalizeTaxKind exige elegir el tipo explícitamente (D4-03): vacío o
// distinto de DEBIT/CREDIT → 400.
func normalizeTaxKind(kind string) (TaxKind, error) {
	switch TaxKind(strings.ToUpper(strings.TrimSpace(kind))) {
	case TaxKindDebit:
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
