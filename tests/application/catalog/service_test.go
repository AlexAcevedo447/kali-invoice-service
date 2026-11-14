package catalog_test

import (
	"errors"
	"fmt"
	"testing"

	appcatalog "github.com/AlexAcevedo447/kali-invoice-service/internal/application/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/domainerr"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/AlexAcevedo447/kali-invoice-service/tests/fakes"
)

func newService() (*appcatalog.Service, *fakes.CustomerRepo, *fakes.ProductRepo) {
	c, p := &fakes.CustomerRepo{}, &fakes.ProductRepo{}
	return appcatalog.NewService(c, p), c, p
}

func TestCreateCustomer_Validation(t *testing.T) {
	svc, repo, _ := newService()
	cases := []struct {
		tenant, name, doc string
		want              error
		validation        bool
	}{
		{"", "ACME", "1", catalog.ErrTenantRequired, false},
		{"t", "  ", "1", catalog.ErrCustomerNameRequired, true},
		{"t", "ACME", " ", catalog.ErrCustomerDocRequired, true},
	}
	for _, tc := range cases {
		_, err := svc.CreateCustomer(tc.tenant, tc.name, tc.doc)
		if !errors.Is(err, tc.want) {
			t.Errorf("(%q,%q,%q): want %v, got %v", tc.tenant, tc.name, tc.doc, tc.want, err)
		}
		if _, ok := domainerr.AsValidation(err); ok != tc.validation {
			t.Errorf("%v: validation=%v, want %v", err, ok, tc.validation)
		}
	}
	if repo.Creates != 0 {
		t.Errorf("repository must not be called on invalid input, calls=%d", repo.Creates)
	}
}

func TestCreateProduct_Validation(t *testing.T) {
	svc, _, repo := newService()
	cases := []struct {
		tenant, code, name string
		price              float64
		want               error
	}{
		{"", "P", "X", 1, catalog.ErrTenantRequired},
		{"t", " ", "X", 1, catalog.ErrProductCodeRequired},
		{"t", "P", "", 1, catalog.ErrProductNameRequired},
		{"t", "P", "X", -0.01, catalog.ErrProductPriceInvalid},
	}
	for _, tc := range cases {
		if _, err := svc.CreateProduct(tc.tenant, tc.code, tc.name, tc.price); !errors.Is(err, tc.want) {
			t.Errorf("want %v, got %v", tc.want, err)
		}
	}
	if repo.Creates != 0 {
		t.Errorf("repository must not be called on invalid input, calls=%d", repo.Creates)
	}
	p, err := svc.CreateProduct("t", " P-0 ", " Gratis ", 0)
	if err != nil || p.Code != "P-0" || p.Name != "Gratis" || p.UnitPrice != 0 || p.TenantID != "t" || p.ID == "" {
		t.Errorf("price 0 must be accepted and fields trimmed: %+v %v", p, err)
	}
}

func TestCreate_DuplicateIsConflictPerTenant(t *testing.T) {
	svc, _, _ := newService()
	if _, err := svc.CreateCustomer("a", "ACME", "900"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateCustomer("a", "Otro", "900"); !errors.Is(err, catalog.ErrCustomerDocDuplicated) {
		t.Errorf("duplicate customer doc: want ErrCustomerDocDuplicated, got %v", err)
	}
	if _, err := svc.CreateCustomer("b", "ACME", "900"); err != nil {
		t.Errorf("same document in another tenant must be allowed: %v", err)
	}

	if _, err := svc.CreateProduct("a", "P1", "Uno", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateProduct("a", "P1", "Otro", 2); !errors.Is(err, catalog.ErrProductCodeDuplicated) {
		t.Errorf("duplicate product code: want ErrProductCodeDuplicated, got %v", err)
	}
	if _, err := svc.CreateProduct("b", "P1", "Uno", 1); err != nil {
		t.Errorf("same code in another tenant must be allowed: %v", err)
	}
}

func TestCreate_PropagatesRepositoryErrors(t *testing.T) {
	svc, crepo, prepo := newService()
	boom := errors.New("db down")
	crepo.CreateErr = fmt.Errorf("insert: %w", boom)
	prepo.CreateErr = boom
	if _, err := svc.CreateCustomer("a", "ACME", "1"); !errors.Is(err, boom) {
		t.Errorf("want wrapped repo error, got %v", err)
	}
	if _, err := svc.CreateProduct("a", "P", "X", 1); !errors.Is(err, boom) {
		t.Errorf("want repo error, got %v", err)
	}
}

func TestListCustomers_ScopedSearchPaginationAndTotal(t *testing.T) {
	svc, _, _ := newService()
	for i := 1; i <= 5; i++ {
		if _, err := svc.CreateCustomer("a", fmt.Sprintf("Acme %d", i), fmt.Sprintf("90%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = svc.CreateCustomer("a", "Beta", "800")
	_, _ = svc.CreateCustomer("b", "Acme B", "901")

	// Solo el tenant propio, con total real.
	all, total, err := svc.ListCustomers("a", ports.CatalogQuery{Pagination: ports.DefaultPagination()})
	if err != nil || total != 6 || len(all) != 6 {
		t.Fatalf("tenant a: len=%d total=%d err=%v", len(all), total, err)
	}
	for _, c := range all {
		if c.TenantID != "a" {
			t.Errorf("leaked customer of tenant %s", c.TenantID)
		}
	}

	// Búsqueda por nombre (case-insensitive) + paginación: el total cuenta todos los coincidentes.
	page2, total, _ := svc.ListCustomers("a", ports.CatalogQuery{Search: "acme", Pagination: ports.Pagination{Page: 2, PageSize: 2}})
	if total != 5 || len(page2) != 2 || page2[0].Name != "Acme 3" || page2[1].Name != "Acme 4" {
		t.Errorf("search+page2: total=%d items=%v", total, names(page2))
	}
	last, total, _ := svc.ListCustomers("a", ports.CatalogQuery{Search: "acme", Pagination: ports.Pagination{Page: 3, PageSize: 2}})
	if total != 5 || len(last) != 1 {
		t.Errorf("last page: total=%d len=%d", total, len(last))
	}
	beyond, total, _ := svc.ListCustomers("a", ports.CatalogQuery{Search: "acme", Pagination: ports.Pagination{Page: 9, PageSize: 2}})
	if total != 5 || len(beyond) != 0 {
		t.Errorf("beyond last page: total=%d len=%d", total, len(beyond))
	}

	// Búsqueda por documento.
	byDoc, total, _ := svc.ListCustomers("a", ports.CatalogQuery{Search: "800", Pagination: ports.DefaultPagination()})
	if total != 1 || len(byDoc) != 1 || byDoc[0].Name != "Beta" {
		t.Errorf("search by document: total=%d items=%v", total, names(byDoc))
	}

	// El documento 901 existe en tenant b pero no debe aparecer en a.
	cross, total, _ := svc.ListCustomers("a", ports.CatalogQuery{Search: "Acme B", Pagination: ports.DefaultPagination()})
	if total != 0 || len(cross) != 0 {
		t.Errorf("cross-tenant search leaked: %v", names(cross))
	}
}

func TestListProducts_ScopedSearchPaginationAndTotal(t *testing.T) {
	svc, _, _ := newService()
	_, _ = svc.CreateProduct("a", "SRV-1", "Consultoría", 100)
	_, _ = svc.CreateProduct("a", "SRV-2", "Soporte", 50)
	_, _ = svc.CreateProduct("a", "HW-1", "Servidor", 900)
	_, _ = svc.CreateProduct("b", "SRV-9", "Ajeno", 1)

	items, total, err := svc.ListProducts("a", ports.CatalogQuery{Search: "srv", Pagination: ports.Pagination{Page: 1, PageSize: 1}})
	if err != nil || total != 2 || len(items) != 1 {
		t.Fatalf("search by code: len=%d total=%d err=%v", len(items), total, err)
	}
	if items[0].TenantID != "a" {
		t.Errorf("leaked product of tenant %s", items[0].TenantID)
	}
	byName, total, _ := svc.ListProducts("a", ports.CatalogQuery{Search: "servidor", Pagination: ports.DefaultPagination()})
	if total != 1 || len(byName) != 1 || byName[0].Code != "HW-1" {
		t.Errorf("search by name: total=%d", total)
	}
	_, totalB, _ := svc.ListProducts("b", ports.CatalogQuery{Pagination: ports.DefaultPagination()})
	if totalB != 1 {
		t.Errorf("tenant b total = %d, want 1", totalB)
	}
}

func names(cs []*catalog.Customer) []string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.Name)
	}
	return out
}
