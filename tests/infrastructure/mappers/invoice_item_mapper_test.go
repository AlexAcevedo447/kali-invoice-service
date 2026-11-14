package mappers_test

import (
	"testing"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
)

// CA-13.2: el orden de los impuestos se persiste en Position (0..n-1).
func TestInvoiceItemToModel_KeepsTaxOrder(t *testing.T) {
	item, err := invoiceitem.New("inv-1", "prod-1", 2, 100, []invoiceitem.TaxInput{
		{Code: "RETEICA", Kind: "CREDIT", Rate: 0.00414},
		{Code: "IVA", Kind: "DEBIT", Rate: 0.19},
		{Code: "INC", Kind: "DEBIT", Rate: 0.08},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	model := mappers.InvoiceItemToModel(item)
	want := []string{"RETEICA", "IVA", "INC"}
	if len(model.Taxes) != len(want) {
		t.Fatalf("expected %d taxes, got %d", len(want), len(model.Taxes))
	}
	for i, code := range want {
		if model.Taxes[i].Code != code || model.Taxes[i].Position != i {
			t.Fatalf("tax %d: got %s/%d, want %s/%d", i, model.Taxes[i].Code, model.Taxes[i].Position, code, i)
		}
	}
}
