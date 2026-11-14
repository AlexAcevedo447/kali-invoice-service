package invoice_test

import (
	"errors"
	"testing"
	"time"

	appcatalog "github.com/AlexAcevedo447/kali-invoice-service/internal/application/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/AlexAcevedo447/kali-invoice-service/tests/fakes"
)

const (
	tenantA = "tenant-a"
	tenantB = "tenant-b"
)

type app struct {
	store     *fakes.InvoiceStore
	publisher *fakes.Publisher
	catalog   *appcatalog.Service
	create    *command.CreateInvoiceCommand
	update    *command.UpdateInvoiceCommand
	pay       *command.PayInvoiceCommand
	cancel    *command.CancelInvoiceCommand
	getByID   *query.GetInvoiceByIDQuery
	getAll    *query.GetAllInvoicesQuery
}

func newApp() *app {
	store := fakes.NewInvoiceStore()
	customers := &fakes.CustomerRepo{}
	products := &fakes.ProductRepo{}
	pub := &fakes.Publisher{}
	return &app{
		store:     store,
		publisher: pub,
		catalog:   appcatalog.NewService(customers, products),
		create:    command.NewCreateInvoiceCommand(store, store, customers, products),
		update:    command.NewUpdateInvoiceCommand(store, store, customers),
		pay:       command.NewPayInvoiceCommand(store, pub),
		cancel:    command.NewCancelInvoiceCommand(store, pub),
		getByID:   query.NewGetInvoiceByIDQuery(store),
		getAll:    query.NewGetAllInvoicesQuery(store),
	}
}

type tenantData struct {
	customer *catalog.Customer
	product  *catalog.Product
}

func (a *app) seedTenant(t *testing.T, tenant string) tenantData {
	t.Helper()
	c, err := a.catalog.CreateCustomer(tenant, "Cliente "+tenant, "DOC-"+tenant)
	if err != nil {
		t.Fatal(err)
	}
	p, err := a.catalog.CreateProduct(tenant, "P-"+tenant, "Producto "+tenant, 100)
	if err != nil {
		t.Fatal(err)
	}
	return tenantData{c, p}
}

func (a *app) createInvoice(t *testing.T, tenant string, d tenantData) *invoice.Invoice {
	t.Helper()
	inv, err := a.create.Execute(command.CreateInvoiceInput{
		TenantID:   tenant,
		CustomerID: d.customer.ID,
		Items:      []command.InvoiceItemInput{{ItemID: d.product.ID, Quantity: 2, UnitPrice: 100}},
	})
	if err != nil {
		t.Fatalf("create invoice for %s: %v", tenant, err)
	}
	if inv.TenantID != tenant {
		t.Fatalf("invoice tenant = %q, want %q", inv.TenantID, tenant)
	}
	return inv
}

func TestTenantIsolation_GetByIDFromOtherTenantIsNotFound(t *testing.T) {
	a := newApp()
	invA := a.createInvoice(t, tenantA, a.seedTenant(t, tenantA))

	if _, err := a.getByID.Execute(tenantB, invA.ID); !errors.Is(err, invoice.ErrNotFound) {
		t.Errorf("tenant B get A's invoice: want ErrNotFound, got %v", err)
	}
	if got, err := a.getByID.Execute(tenantA, invA.ID); err != nil || got.ID != invA.ID {
		t.Errorf("tenant A must see its own invoice: %v", err)
	}
}

func TestTenantIsolation_ListReturnsOnlyOwnInvoices(t *testing.T) {
	a := newApp()
	dA, dB := a.seedTenant(t, tenantA), a.seedTenant(t, tenantB)
	a.createInvoice(t, tenantA, dA)
	a.createInvoice(t, tenantA, dA)
	invB := a.createInvoice(t, tenantB, dB)

	listA, err := a.getAll.Execute(tenantA, ports.DefaultPagination())
	if err != nil {
		t.Fatal(err)
	}
	if len(listA) != 2 {
		t.Errorf("tenant A list len = %d, want 2", len(listA))
	}
	for _, inv := range listA {
		if inv.TenantID != tenantA || inv.ID == invB.ID {
			t.Errorf("tenant A list leaked invoice %s of tenant %s", inv.ID, inv.TenantID)
		}
	}
	if n, _ := a.getAll.Total(tenantA); n != 2 {
		t.Errorf("tenant A total = %d, want 2", n)
	}
	listB, _ := a.getAll.Execute(tenantB, ports.DefaultPagination())
	if len(listB) != 1 || listB[0].ID != invB.ID {
		t.Errorf("tenant B list = %v, want only %s", listB, invB.ID)
	}
	if n, _ := a.getAll.Total(tenantB); n != 1 {
		t.Errorf("tenant B total = %d, want 1", n)
	}
	if list, _ := a.getAll.Execute("tenant-c", ports.DefaultPagination()); len(list) != 0 {
		t.Errorf("unknown tenant must see nothing, got %d", len(list))
	}
}

func TestTenantIsolation_UpdatePayCancelFromOtherTenantAreNotFound(t *testing.T) {
	a := newApp()
	dA, dB := a.seedTenant(t, tenantA), a.seedTenant(t, tenantB)
	invA := a.createInvoice(t, tenantA, dA)
	before, _ := a.store.Raw(invA.ID)

	due := time.Now().AddDate(0, 2, 0)
	if _, err := a.update.Execute(tenantB, invA.ID, command.UpdateInvoiceInput{CustomerID: &dB.customer.ID, DueDate: &due}); !errors.Is(err, invoice.ErrNotFound) {
		t.Errorf("update from B: want ErrNotFound, got %v", err)
	}
	if err := a.pay.Execute(tenantB, invA.ID); !errors.Is(err, invoice.ErrNotFound) {
		t.Errorf("pay from B: want ErrNotFound, got %v", err)
	}
	if err := a.cancel.Execute(tenantB, invA.ID); !errors.Is(err, invoice.ErrNotFound) {
		t.Errorf("cancel from B: want ErrNotFound, got %v", err)
	}

	after, _ := a.store.Raw(invA.ID)
	if after.Status != invoice.StatusPending || after.CustomerID != before.CustomerID || !after.DueDate.Equal(before.DueDate) || after.TenantID != tenantA {
		t.Errorf("tenant A's invoice was modified by tenant B: before=%+v after=%+v", before, after)
	}
	if len(a.publisher.Events) != 0 {
		t.Errorf("no status events expected, got %d", len(a.publisher.Events))
	}

	// El dueño sí puede operar su factura.
	if err := a.pay.Execute(tenantA, invA.ID); err != nil {
		t.Errorf("tenant A pay: %v", err)
	}
}

func TestTenantIsolation_CreateWithForeignCustomerOrProduct(t *testing.T) {
	a := newApp()
	dA, dB := a.seedTenant(t, tenantA), a.seedTenant(t, tenantB)

	_, err := a.create.Execute(command.CreateInvoiceInput{
		TenantID:   tenantA,
		CustomerID: dB.customer.ID,
		Items:      []command.InvoiceItemInput{{ItemID: dA.product.ID, Quantity: 1, UnitPrice: 10}},
	})
	if !errors.Is(err, catalog.ErrCustomerNotFound) {
		t.Errorf("foreign customer: want ErrCustomerNotFound, got %v", err)
	}

	_, err = a.create.Execute(command.CreateInvoiceInput{
		TenantID:   tenantA,
		CustomerID: dA.customer.ID,
		Items: []command.InvoiceItemInput{
			{ItemID: dA.product.ID, Quantity: 1, UnitPrice: 10},
			{ItemID: dB.product.ID, Quantity: 1, UnitPrice: 10},
		},
	})
	if !errors.Is(err, catalog.ErrProductNotFound) {
		t.Errorf("foreign product: want ErrProductNotFound, got %v", err)
	}

	if n, _ := a.getAll.Total(tenantA); n != 0 {
		t.Errorf("no invoice must be persisted, total=%d", n)
	}
}

func TestTenantIsolation_UpdateWithForeignCustomerIsRejected(t *testing.T) {
	a := newApp()
	dA, dB := a.seedTenant(t, tenantA), a.seedTenant(t, tenantB)
	invA := a.createInvoice(t, tenantA, dA)

	if _, err := a.update.Execute(tenantA, invA.ID, command.UpdateInvoiceInput{CustomerID: &dB.customer.ID}); !errors.Is(err, catalog.ErrCustomerNotFound) {
		t.Errorf("want ErrCustomerNotFound, got %v", err)
	}
	if after, _ := a.store.Raw(invA.ID); after.CustomerID != dA.customer.ID {
		t.Errorf("customer changed to foreign one: %s", after.CustomerID)
	}
}
