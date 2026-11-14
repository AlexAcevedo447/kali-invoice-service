package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appcatalog "github.com/AlexAcevedo447/kali-invoice-service/internal/application/catalog"
	appinvoicecmd "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	appinvoicequery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	appitemcmd "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	httpinfra "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/tests/fakes"
	"github.com/gofiber/fiber/v2"
)

const (
	tenantA = "tenant-a"
	tenantB = "tenant-b"
)

// itemStore es un fake mínimo de InvoiceItemByIDReader/InvoiceItemUpdater.
type itemStore struct {
	tenant string
	item   *invoiceitem.InvoiceItem
}

func (s *itemStore) GetByID(tenantID, id string) (*invoiceitem.InvoiceItem, error) {
	if s.item == nil || s.item.ID != id || tenantID != s.tenant {
		return nil, invoiceitem.ErrNotFound
	}
	cp := *s.item
	return &cp, nil
}

func (s *itemStore) Update(tenantID string, item *invoiceitem.InvoiceItem) error {
	if tenantID != s.tenant {
		return invoiceitem.ErrNotFound
	}
	s.item = item
	return nil
}

type env struct {
	app       *fiber.App
	auth      *fakes.Auth
	invoices  *fakes.InvoiceStore
	customers *fakes.CustomerRepo
	products  *fakes.ProductRepo
	items     *itemStore
	svc       *appcatalog.Service
}

// newEnv monta los handlers reales sobre la app Fiber real (con su
// ErrorHandler central) y el middleware de autenticación real.
func newEnv(t *testing.T) *env {
	t.Helper()
	e := &env{
		auth:      fakes.NewAuth(t),
		invoices:  fakes.NewInvoiceStore(),
		customers: &fakes.CustomerRepo{},
		products:  &fakes.ProductRepo{},
		items:     &itemStore{tenant: tenantA},
	}
	e.svc = appcatalog.NewService(e.customers, e.products)
	pub := &fakes.Publisher{}
	getByID := appinvoicequery.NewGetInvoiceByIDQuery(e.invoices)

	catalogH := handlers.NewCatalogHandler(e.svc)
	createH := handlers.NewCreateInvoiceHandler(appinvoicecmd.NewCreateInvoiceCommand(e.invoices, e.invoices, e.customers, e.products))
	updateH := handlers.NewUpdateInvoiceHandler(appinvoicecmd.NewUpdateInvoiceCommand(e.invoices, e.invoices, e.customers))
	getH := handlers.NewGetInvoiceByIDHandler(getByID)
	payH := handlers.NewPayInvoiceHandler(appinvoicecmd.NewPayInvoiceCommand(e.invoices, pub), getByID)
	cancelH := handlers.NewCancelInvoiceHandler(appinvoicecmd.NewCancelInvoiceCommand(e.invoices, pub), getByID)
	updateItemH := handlers.NewUpdateInvoiceItemHandler(appitemcmd.NewUpdateInvoiceItemCommand(e.items, e.items))

	app := httpinfra.NewFiberApp()
	inv := app.Group("/invoices", e.auth.Handler)
	inv.Post("/", createH.Handle)
	inv.Put("/:id", updateH.Handle)
	inv.Get("/:id", getH.Handle)
	inv.Patch("/:id/pay", payH.Handle)
	inv.Patch("/:id/cancel", cancelH.Handle)
	app.Put("/invoice-items/:id", e.auth.Handler, updateItemH.Handle)
	cu := app.Group("/customers", e.auth.Handler)
	cu.Post("/", catalogH.CreateCustomer)
	pr := app.Group("/products", e.auth.Handler)
	pr.Post("/", catalogH.CreateProduct)
	e.app = app
	return e
}

func (e *env) do(t *testing.T, tenant, method, path, body string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", e.auth.Bearer(t, tenant))
	resp, err := e.app.Test(req, 10000)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(raw, &parsed)
	return resp.StatusCode, parsed.Error
}

func expect(t *testing.T, gotStatus int, gotMsg string, wantStatus int, wantMsg string) {
	t.Helper()
	if gotStatus != wantStatus || gotMsg != wantMsg {
		t.Errorf("got %d %q, want %d %q", gotStatus, gotMsg, wantStatus, wantMsg)
	}
}

// Cada error (directo o envuelto) que sale del caso de uso se traduce al
// código HTTP esperado, con el texto exacto del error de dominio.
func TestErrorMapping_StatusByError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"invoice not found", invoice.ErrNotFound, 404, "invoice not found"},
		{"invoice item not found", invoiceitem.ErrNotFound, 404, "invoice item not found"},
		{"already paid", invoice.ErrAlreadyPaid, 422, "invoice is already paid"},
		{"already canceled", invoice.ErrAlreadyCanceled, 422, "invoice is already canceled"},
		{"customer not found", catalog.ErrCustomerNotFound, 422, "customer not found"},
		{"product not found", catalog.ErrProductNotFound, 422, "product not found"},
		{"duplicate customer", catalog.ErrCustomerDocDuplicated, 409, "a customer with this document already exists"},
		{"duplicate product", catalog.ErrProductCodeDuplicated, 409, "a product with this code already exists"},
		{"tenant required", invoice.ErrTenantRequired, 401, "tenant is required"},
		{"validation customer_id", invoice.ErrCustomerIDRequired, 400, "customer_id is required"},
		{"validation due_date", invoice.ErrDueDateBeforeIssueDate, 400, "due_date cannot be before issue_date"},
		{"validation items", invoice.ErrItemsRequired, 400, "invoice must have at least one item"},
		{"validation quantity", invoiceitem.ErrQuantityInvalid, 400, "quantity must be greater than 0"},
		{"validation unit_price", invoiceitem.ErrUnitPriceInvalid, 400, "unit_price must be greater than 0"},
		{"validation tax code", invoiceitem.ErrTaxCodeRequired, 400, "tax code is required"},
		{"validation tax rate", invoiceitem.ErrTaxRateInvalid, 400, "tax rate must be greater than or equal to 0"},
		{"validation tax kind", invoiceitem.ErrTaxKindInvalid, 400, "tax kind must be DEBIT or CREDIT"},
		{"validation catalog", catalog.ErrProductPriceInvalid, 400, "product unit_price must be greater than or equal to 0"},
		{"wrapped not found", fmt.Errorf("repo: %w", invoice.ErrNotFound), 404, "invoice not found"},
		{"wrapped already paid", fmt.Errorf("tx: %w", invoice.ErrAlreadyPaid), 422, "invoice is already paid"},
		{"wrapped duplicate", fmt.Errorf("insert: %w", catalog.ErrCustomerDocDuplicated), 409, "a customer with this document already exists"},
		{"double wrapped product", fmt.Errorf("a: %w", fmt.Errorf("b: %w", catalog.ErrProductNotFound)), 422, "product not found"},
		{"wrapped validation", fmt.Errorf("update: %w", invoice.ErrDueDateBeforeIssueDate), 400, "due_date cannot be before issue_date"},
		{"joined validation", errors.Join(errors.New("ctx"), invoiceitem.ErrQuantityInvalid), 400, "quantity must be greater than 0"},
		// Lo inesperado responde un mensaje genérico: el detalle interno solo va al log.
		{"unexpected", errors.New("connection refused"), 500, "internal server error"},
		{"unexpected sql detail", errors.New(`pq: relation "invoices" does not exist (host=db user=kali_dev)`), 500, "internal server error"},
		{"wrapped unexpected", fmt.Errorf("list: %w", errors.New("dial tcp 10.0.0.5:5432: i/o timeout")), 500, "internal server error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t)
			// El repositorio falla con el error a mapear; el handler real lo traduce.
			e.customers.CreateErr = tc.err
			status, msg := e.do(t, tenantA, "POST", "/customers/", `{"name":"ACME","document":"900"}`)
			expect(t, status, msg, tc.wantStatus, tc.wantMsg)
		})
	}
}

// Flujos reales: el error nace en el dominio y atraviesa caso de uso + handler.
func TestErrorMapping_EndToEndThroughHandlers(t *testing.T) {
	e := newEnv(t)

	// Catálogo: validación de dominio → 400 (antes también eran 400; deben seguir así).
	s, m := e.do(t, tenantA, "POST", "/customers/", `{"name":"  ","document":"1"}`)
	expect(t, s, m, 400, "customer name is required")
	s, m = e.do(t, tenantA, "POST", "/products/", `{"code":"P1","name":"X","unit_price":-1}`)
	expect(t, s, m, 400, "product unit_price must be greater than or equal to 0")

	// Duplicado real dentro del tenant → 409; en otro tenant no colisiona.
	s, _ = e.do(t, tenantA, "POST", "/customers/", `{"name":"ACME","document":"900"}`)
	expect(t, s, "", 201, "")
	s, m = e.do(t, tenantA, "POST", "/customers/", `{"name":"ACME 2","document":"900"}`)
	expect(t, s, m, 409, "a customer with this document already exists")
	s, _ = e.do(t, tenantB, "POST", "/customers/", `{"name":"ACME","document":"900"}`)
	expect(t, s, "", 201, "")

	// Factura de tenant A.
	cust, _ := e.svc.CreateCustomer(tenantA, "Cliente", "C-1")
	prod, _ := e.svc.CreateProduct(tenantA, "P-1", "Producto", 10)
	issue := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	inv, err := invoice.New(tenantA, cust.ID, issue, issue.AddDate(0, 0, 30), []invoice.ItemInput{{ItemID: prod.ID, Quantity: 1, UnitPrice: 10}})
	if err != nil {
		t.Fatal(err)
	}
	_ = e.invoices.Save(inv)
	path := "/invoices/" + inv.ID

	// Validaciones que antes caían en 500.
	s, m = e.do(t, tenantA, "PUT", path, `{"due_date":"2026-01-01T00:00:00Z"}`)
	expect(t, s, m, 400, "due_date cannot be before issue_date")
	s, m = e.do(t, tenantA, "PUT", path, `{"customer_id":""}`)
	expect(t, s, m, 400, "customer_id is required")

	// Cliente de otro tenant → 422.
	other, _ := e.svc.CreateCustomer(tenantB, "Ajeno", "B-1")
	s, m = e.do(t, tenantA, "PUT", path, `{"customer_id":"`+other.ID+`"}`)
	expect(t, s, m, 422, "customer not found")

	// Creación con producto de otro tenant → 422.
	otherProd, _ := e.svc.CreateProduct(tenantB, "P-B", "Ajeno", 5)
	body := fmt.Sprintf(`{"customer_id":%q,"items":[{"item_id":%q,"quantity":1,"unit_price":5}]}`, cust.ID, otherProd.ID)
	s, m = e.do(t, tenantA, "POST", "/invoices/", body)
	expect(t, s, m, 422, "product not found")

	// Validación de ítem en creación de factura → 400.
	body = fmt.Sprintf(`{"customer_id":%q,"items":[{"item_id":%q,"quantity":0,"unit_price":5}]}`, cust.ID, prod.ID)
	s, m = e.do(t, tenantA, "POST", "/invoices/", body)
	expect(t, s, m, 400, "quantity must be greater than 0")

	// Otro tenant no ve la factura → 404.
	s, m = e.do(t, tenantB, "GET", path, "")
	expect(t, s, m, 404, "invoice not found")
	s, m = e.do(t, tenantB, "PATCH", path+"/pay", "")
	expect(t, s, m, 404, "invoice not found")

	// Pagar dos veces → 422; cancelar pagada → 422; editar pagada → 422.
	s, _ = e.do(t, tenantA, "PATCH", path+"/pay", "")
	expect(t, s, "", 200, "")
	s, m = e.do(t, tenantA, "PATCH", path+"/pay", "")
	expect(t, s, m, 422, "invoice is already paid")
	s, m = e.do(t, tenantA, "PATCH", path+"/cancel", "")
	expect(t, s, m, 422, "invoice is already paid")
	s, m = e.do(t, tenantA, "PUT", path, `{"due_date":"2026-03-01T00:00:00Z"}`)
	expect(t, s, m, 422, "invoice is already paid")

	// Ítems: validaciones de invoiceitem → 400.
	item, err := invoiceitem.New(inv.ID, prod.ID, 1, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	e.items.item = item
	s, m = e.do(t, tenantA, "PUT", "/invoice-items/"+item.ID, `{"quantity":0,"unit_price":10}`)
	expect(t, s, m, 400, "quantity must be greater than 0")
	s, m = e.do(t, tenantA, "PUT", "/invoice-items/"+item.ID, `{"quantity":1,"unit_price":10,"taxes":[{"code":"IVA","kind":"X","rate":0.19}]}`)
	expect(t, s, m, 400, "tax kind must be DEBIT or CREDIT")
	s, m = e.do(t, tenantB, "PUT", "/invoice-items/"+item.ID, `{"quantity":1,"unit_price":10}`)
	expect(t, s, m, 404, "invoice item not found")
}
