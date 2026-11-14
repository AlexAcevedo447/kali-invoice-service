package invoiceitem_test

import (
	"errors"
	"math"
	"testing"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
)

func TestNew_TaxAmountAlwaysPositive_AndNatureAffectsNetTax(t *testing.T) {
	item, err := invoiceitem.New("inv-1", "item-1", 2, 100, []invoiceitem.TaxInput{
		{Code: "IVA", Kind: "DEBIT", Rate: 0.19},
		{Code: "RETEFUENTE", Kind: "CREDIT", Rate: 0.025},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(item.Taxes) != 2 {
		t.Fatalf("expected 2 taxes, got %d", len(item.Taxes))
	}

	for _, tx := range item.Taxes {
		if tx.Amount < 0 {
			t.Fatalf("expected positive tax amount, got %f", tx.Amount)
		}
	}

	if math.Abs(item.TaxTotal-33) > 0.000001 {
		t.Fatalf("expected net tax total 33, got %f", item.TaxTotal)
	}
	if math.Abs(item.Total-233) > 0.000001 {
		t.Fatalf("expected line total 233, got %f", item.Total)
	}
}

// D4-03: el tipo se elige explícitamente; vacío ya no se asume DEBIT.
func TestNew_EmptyKindRejected(t *testing.T) {
	_, err := invoiceitem.New("inv-1", "item-1", 1, 100, []invoiceitem.TaxInput{
		{Code: "IVA", Rate: 0.19},
	})
	if !errors.Is(err, invoiceitem.ErrTaxKindInvalid) {
		t.Fatalf("expected ErrTaxKindInvalid for empty kind, got %v", err)
	}
}

// D4-02: un código no se repite en la línea (sin distinguir mayúsculas ni espacios).
func TestNew_DuplicatedTaxCodeRejected(t *testing.T) {
	_, err := invoiceitem.New("inv-1", "item-1", 1, 100, []invoiceitem.TaxInput{
		{Code: "IVA", Kind: "DEBIT", Rate: 0.19},
		{Code: " iva ", Kind: "CREDIT", Rate: 0.05},
	})
	if !errors.Is(err, invoiceitem.ErrTaxCodeDuplicated) {
		t.Fatalf("expected ErrTaxCodeDuplicated, got %v", err)
	}
}

// D4-03: código normalizado (sin espacios exteriores, en mayúsculas); tipo sin distinguir mayúsculas.
func TestNew_TaxCodeAndKindNormalized(t *testing.T) {
	item, err := invoiceitem.New("inv-1", "item-1", 1, 100, []invoiceitem.TaxInput{
		{Code: "  reteFuente ", Kind: "credit", Rate: 0.025},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Taxes[0].Code != "RETEFUENTE" || item.Taxes[0].Kind != invoiceitem.TaxKindCredit {
		t.Fatalf("expected RETEFUENTE/CREDIT, got %s/%s", item.Taxes[0].Code, item.Taxes[0].Kind)
	}
}

// Código vacío o solo espacios → 400.
func TestNew_BlankTaxCodeRejected(t *testing.T) {
	_, err := invoiceitem.New("inv-1", "item-1", 1, 100, []invoiceitem.TaxInput{{Code: "   ", Kind: "DEBIT", Rate: 0.1}})
	if !errors.Is(err, invoiceitem.ErrTaxCodeRequired) {
		t.Fatalf("expected ErrTaxCodeRequired, got %v", err)
	}
}

// D4-01: una línea sin impuestos es válida (impuestos 0, total = subtotal).
func TestNew_NoTaxesAllowed(t *testing.T) {
	item, err := invoiceitem.New("inv-1", "item-1", 2, 50, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(item.Taxes) != 0 || item.TaxTotal != 0 || item.Total != 100 {
		t.Fatalf("expected no taxes, tax total 0 and total 100, got %d taxes, %f, %f", len(item.Taxes), item.TaxTotal, item.Total)
	}
}

// Varios impuestos DEBIT y CREDIT con tarifas decimales: el servidor no redondea
// (importe = subtotal × tarifa en float64; débito suma, crédito resta).
func TestNew_MultipleDecimalTaxes_NoRounding(t *testing.T) {
	item, err := invoiceitem.New("inv-1", "item-1", 3, 33.33, []invoiceitem.TaxInput{
		{Code: "IVA", Kind: "DEBIT", Rate: 0.19},
		{Code: "INC", Kind: "DEBIT", Rate: 0.08},
		{Code: "RETEFUENTE", Kind: "CREDIT", Rate: 0.025},
		{Code: "RETEICA", Kind: "CREDIT", Rate: 0.00414},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	subtotal := 3 * 33.33
	want := subtotal*0.19 + subtotal*0.08 - subtotal*0.025 - subtotal*0.00414
	if len(item.Taxes) != 4 {
		t.Fatalf("expected 4 taxes, got %d", len(item.Taxes))
	}
	if item.TaxTotal != want {
		t.Fatalf("tax total must be the exact float64 sum without rounding: want %v got %v", want, item.TaxTotal)
	}
	if item.Taxes[3].Amount != subtotal*0.00414 {
		t.Fatalf("amount must be subtotal × rate without rounding, got %v", item.Taxes[3].Amount)
	}
}

func TestNew_InvalidTaxKind(t *testing.T) {
	_, err := invoiceitem.New("inv-1", "item-1", 1, 100, []invoiceitem.TaxInput{
		{Code: "X", Kind: "WHATEVER", Rate: 0.1},
	})
	if err == nil {
		t.Fatal("expected error for invalid tax kind")
	}
}
