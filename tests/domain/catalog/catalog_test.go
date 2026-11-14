package catalog_test

import (
	"testing"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
)

func TestNewCustomer_GeneratesIDAndKeepsTenant(t *testing.T) {
	c, err := catalog.NewCustomer("tenant-a", "  ACME S.A.S. ", " 900123456 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID == "" || c.TenantID != "tenant-a" || c.Name != "ACME S.A.S." || c.Document != "900123456" {
		t.Errorf("unexpected customer: %+v", c)
	}
}

func TestNewCustomer_Validation(t *testing.T) {
	cases := map[string]struct {
		tenant, name, doc string
		want              error
	}{
		"sin tenant":    {"", "ACME", "1", catalog.ErrTenantRequired},
		"sin nombre":    {"t", " ", "1", catalog.ErrCustomerNameRequired},
		"sin documento": {"t", "ACME", "", catalog.ErrCustomerDocRequired},
	}
	for name, tc := range cases {
		if _, err := catalog.NewCustomer(tc.tenant, tc.name, tc.doc); err != tc.want {
			t.Errorf("%s: expected %v, got %v", name, tc.want, err)
		}
	}
}

func TestNewProduct_Validation(t *testing.T) {
	if _, err := catalog.NewProduct("t", "", "Servicio", 10); err != catalog.ErrProductCodeRequired {
		t.Errorf("expected ErrProductCodeRequired, got %v", err)
	}
	if _, err := catalog.NewProduct("t", "SRV-1", "", 10); err != catalog.ErrProductNameRequired {
		t.Errorf("expected ErrProductNameRequired, got %v", err)
	}
	if _, err := catalog.NewProduct("t", "SRV-1", "Servicio", -1); err != catalog.ErrProductPriceInvalid {
		t.Errorf("expected ErrProductPriceInvalid, got %v", err)
	}
	p, err := catalog.NewProduct("t", "SRV-1", "Servicio", 0)
	if err != nil || p.ID == "" || p.TenantID != "t" {
		t.Errorf("unexpected product %+v err %v", p, err)
	}
}
