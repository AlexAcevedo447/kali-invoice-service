package invoiceitem_test

import (
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

func TestNew_EmptyKindDefaultsToDebit(t *testing.T) {
	item, err := invoiceitem.New("inv-1", "item-1", 1, 100, []invoiceitem.TaxInput{
		{Code: "IVA", Rate: 0.19},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := item.Taxes[0].Kind; got != invoiceitem.TaxKindDebit {
		t.Fatalf("expected default kind DEBIT, got %s", got)
	}
	if math.Abs(item.TaxTotal-19) > 0.000001 {
		t.Fatalf("expected tax total 19, got %f", item.TaxTotal)
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
