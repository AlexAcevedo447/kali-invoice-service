package routes_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	appcatalog "github.com/AlexAcevedo447/kali-invoice-service/internal/application/catalog"
	appinvoicecmd "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	appinvoicequery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	appitemcmd "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/command"
	appitemquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	httpinfra "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/routes"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/AlexAcevedo447/kali-invoice-service/tests/fakes"
	"github.com/gofiber/fiber/v2"
)

const tenant = "tenant-a"

// itemRepo es un repositorio de ítems en memoria mínimo (vacío por defecto).
type itemRepo struct {
	mu    sync.Mutex
	items map[string]invoiceitem.InvoiceItem
}

func (r *itemRepo) Save(tenantID string, item *invoiceitem.InvoiceItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[tenantID+"|"+item.ID] = *item
	return nil
}

func (r *itemRepo) GetAll(tenantID string, _ ports.Pagination) ([]*invoiceitem.InvoiceItem, error) {
	return []*invoiceitem.InvoiceItem{}, nil
}

func (r *itemRepo) GetByID(tenantID, id string) (*invoiceitem.InvoiceItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	it, ok := r.items[tenantID+"|"+id]
	if !ok {
		return nil, invoiceitem.ErrNotFound
	}
	return &it, nil
}

func (r *itemRepo) Update(tenantID string, item *invoiceitem.InvoiceItem) error {
	return r.Save(tenantID, item)
}

func (r *itemRepo) DeleteByID(tenantID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, tenantID+"|"+id)
	return nil
}

type env struct {
	app     *fiber.App
	auth    *fakes.Auth
	session *fakes.Session
}

// newEnv monta las rutas reales (routes.RegisterDomainRoutes) bajo /api/v1 con
// el middleware de autenticación real, la comprobación de sesión contra un
// /auth/session simulado y repositorios en memoria.
func newEnv(t *testing.T, ttl time.Duration) *env {
	t.Helper()
	e := &env{session: fakes.NewSession(t)}
	e.auth = fakes.NewAuthWithStatus(t, middleware.NewTenantStatusChecker(e.session.URL, ttl))

	invoices := fakes.NewInvoiceStore()
	customers := &fakes.CustomerRepo{}
	products := &fakes.ProductRepo{}
	items := &itemRepo{items: map[string]invoiceitem.InvoiceItem{}}
	pub := &fakes.Publisher{}
	getByID := appinvoicequery.NewGetInvoiceByIDQuery(invoices)

	app := httpinfra.NewFiberApp()
	routes.RegisterDomainRoutes(
		app.Group("/api/v1"),
		handlers.NewCreateInvoiceHandler(appinvoicecmd.NewCreateInvoiceCommand(invoices, invoices, customers, products)),
		handlers.NewUpdateInvoiceHandler(appinvoicecmd.NewUpdateInvoiceCommand(invoices, invoices, customers)),
		handlers.NewGetAllInvoicesHandler(appinvoicequery.NewGetAllInvoicesQuery(invoices)),
		handlers.NewGetInvoiceByIDHandler(getByID),
		handlers.NewPayInvoiceHandler(appinvoicecmd.NewPayInvoiceCommand(invoices, pub), getByID),
		handlers.NewCancelInvoiceHandler(appinvoicecmd.NewCancelInvoiceCommand(invoices, pub), getByID),
		handlers.NewCreateInvoiceItemHandler(appitemcmd.NewCreateInvoiceItemCommand(items, products)),
		handlers.NewGetAllInvoiceItemsHandler(appitemquery.NewGetAllInvoiceItemsQuery(items)),
		handlers.NewGetInvoiceItemByIDHandler(appitemquery.NewGetInvoiceItemByIDQuery(items)),
		handlers.NewUpdateInvoiceItemHandler(appitemcmd.NewUpdateInvoiceItemCommand(items, items)),
		handlers.NewDeleteInvoiceItemHandler(appitemcmd.NewDeleteInvoiceItemCommand(items)),
		fakes.NewIdempotencyRepo(),
		handlers.NewCatalogHandler(appcatalog.NewService(customers, products)),
		e.auth.Handler,
	)
	e.app = app
	return e
}

var keySeq atomic.Int64

// do envía la petición con el token del usuario (bearer vacío = sin token).
func (e *env) do(t *testing.T, bearer, method, path, body string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", fmt.Sprintf("key-%d", keySeq.Add(1)))
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
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

type route struct{ method, path, body string }

// protected son todas las rutas de negocio con el permiso que exigen según el
// contrato (MATRIZ_PERMISOS.md, sección kali-invoice-service).
var protected = []struct {
	route
	resource, action string
}{
	{route{http.MethodGet, "/api/v1/invoices", ""}, "invoices", "read"},
	{route{http.MethodGet, "/api/v1/invoices/00000000-0000-4000-8000-000000000001", ""}, "invoices", "read"},
	{route{http.MethodPost, "/api/v1/invoices", `{}`}, "invoices", "manage"},
	{route{http.MethodPut, "/api/v1/invoices/00000000-0000-4000-8000-000000000001", `{}`}, "invoices", "manage"},
	{route{http.MethodPatch, "/api/v1/invoices/00000000-0000-4000-8000-000000000001/pay", ""}, "invoices", "manage"},
	{route{http.MethodPatch, "/api/v1/invoices/00000000-0000-4000-8000-000000000001/cancel", ""}, "invoices", "manage"},
	{route{http.MethodGet, "/api/v1/invoice-items", ""}, "invoice-items", "read"},
	{route{http.MethodGet, "/api/v1/invoice-items/00000000-0000-4000-8000-000000000002", ""}, "invoice-items", "read"},
	{route{http.MethodPost, "/api/v1/invoice-items", `{}`}, "invoice-items", "manage"},
	{route{http.MethodPut, "/api/v1/invoice-items/00000000-0000-4000-8000-000000000002", `{}`}, "invoice-items", "manage"},
	{route{http.MethodDelete, "/api/v1/invoice-items/00000000-0000-4000-8000-000000000002", ""}, "invoice-items", "manage"},
	{route{http.MethodGet, "/api/v1/customers", ""}, "customers", "read"},
	{route{http.MethodPost, "/api/v1/customers", `{}`}, "customers", "manage"},
	{route{http.MethodGet, "/api/v1/products", ""}, "products", "read"},
	{route{http.MethodPost, "/api/v1/products", `{}`}, "products", "manage"},
}

func expectDenied(t *testing.T, status int, msg string, r route) {
	t.Helper()
	if status != http.StatusForbidden || msg != "insufficient permissions" {
		t.Errorf("%s %s: expected 403 insufficient permissions, got %d %q", r.method, r.path, status, msg)
	}
}

func expectAllowed(t *testing.T, status int, msg string, r route) {
	t.Helper()
	if status == http.StatusForbidden || status == http.StatusUnauthorized || status >= 500 {
		t.Errorf("%s %s: expected to pass authorization, got %d %q", r.method, r.path, status, msg)
	}
}

func TestRoutes_InvoicesRead_CanListButNotWrite(t *testing.T) {
	e := newEnv(t, 0)
	e.session.Grant(tenant, "reader", "invoices/read")
	bearer := e.auth.BearerFor(t, tenant, "reader")

	if status, msg := e.do(t, bearer, http.MethodGet, "/api/v1/invoices", ""); status != http.StatusOK {
		t.Fatalf("GET /invoices: expected 200, got %d %q", status, msg)
	}
	// Leer una factura inexistente pasa la autorización y responde 404.
	if status, msg := e.do(t, bearer, http.MethodGet, "/api/v1/invoices/00000000-0000-4000-8000-000000000001", ""); status != http.StatusNotFound {
		t.Fatalf("GET /invoices/:id: expected 404, got %d %q", status, msg)
	}
	for _, r := range []route{
		{http.MethodPost, "/api/v1/invoices", `{}`},
		{http.MethodPut, "/api/v1/invoices/00000000-0000-4000-8000-000000000001", `{}`},
		{http.MethodPatch, "/api/v1/invoices/00000000-0000-4000-8000-000000000001/pay", ""},
		{http.MethodPatch, "/api/v1/invoices/00000000-0000-4000-8000-000000000001/cancel", ""},
	} {
		status, msg := e.do(t, bearer, r.method, r.path, r.body)
		expectDenied(t, status, msg, r)
	}
	// Tampoco concede otros recursos.
	status, msg := e.do(t, bearer, http.MethodGet, "/api/v1/customers", "")
	expectDenied(t, status, msg, route{http.MethodGet, "/api/v1/customers", ""})
}

func TestRoutes_InvoicesManage_ImpliesRead(t *testing.T) {
	e := newEnv(t, 0)
	e.session.Grant(tenant, "manager", "invoices/manage")
	bearer := e.auth.BearerFor(t, tenant, "manager")
	for _, p := range protected {
		status, msg := e.do(t, bearer, p.method, p.path, p.body)
		if p.resource == "invoices" {
			expectAllowed(t, status, msg, p.route)
		} else {
			expectDenied(t, status, msg, p.route)
		}
	}
}

func TestRoutes_CustomersRead_CanListButNotCreate(t *testing.T) {
	e := newEnv(t, 0)
	e.session.Grant(tenant, "clerk", "customers/read")
	bearer := e.auth.BearerFor(t, tenant, "clerk")

	if status, msg := e.do(t, bearer, http.MethodGet, "/api/v1/customers", ""); status != http.StatusOK {
		t.Fatalf("GET /customers: expected 200, got %d %q", status, msg)
	}
	status, msg := e.do(t, bearer, http.MethodPost, "/api/v1/customers", `{"name":"ACME","document":"900"}`)
	expectDenied(t, status, msg, route{http.MethodPost, "/api/v1/customers", ""})
}

func TestRoutes_EachRouteRequiresItsPermission(t *testing.T) {
	e := newEnv(t, 0)
	for i, p := range protected {
		user := "user-" + p.resource + "-" + p.action
		e.session.Grant(tenant, user, p.resource+"/"+p.action)
		bearer := e.auth.BearerFor(t, tenant, user)
		status, msg := e.do(t, bearer, p.method, p.path, p.body)
		expectAllowed(t, status, msg, p.route)
		// El mismo usuario no alcanza las rutas que exigen otro permiso.
		for j, other := range protected {
			if i == j || other.resource == p.resource && (other.action == p.action || p.action == "manage") {
				continue
			}
			status, msg := e.do(t, bearer, other.method, other.path, other.body)
			expectDenied(t, status, msg, other.route)
		}
	}
}

func TestRoutes_WildcardGrantsEverything(t *testing.T) {
	e := newEnv(t, 0)
	e.session.Grant(tenant, "admin", "*/*")
	bearer := e.auth.BearerFor(t, tenant, "admin")
	for _, p := range protected {
		status, msg := e.do(t, bearer, p.method, p.path, p.body)
		expectAllowed(t, status, msg, p.route)
	}
}

func TestRoutes_NoPermissions_DeniedEverywhereExceptMetrics(t *testing.T) {
	e := newEnv(t, 0)
	bearer := e.auth.BearerFor(t, tenant, "nobody") // sesión activa, lista vacía
	for _, p := range protected {
		status, msg := e.do(t, bearer, p.method, p.path, p.body)
		expectDenied(t, status, msg, p.route)
	}
	if status, _ := e.do(t, "", http.MethodGet, "/api/v1/metrics", ""); status != http.StatusOK {
		t.Fatalf("GET /metrics without token: expected 200, got %d", status)
	}
}

// Sin comprobación de sesión configurada (TenantStatus nil) no hay permisos:
// las rutas fallan cerradas.
func TestRoutes_WithoutSessionCheck_FailsClosed(t *testing.T) {
	auth := fakes.NewAuth(t) // autenticación sin TenantStatus
	app := httpinfra.NewFiberApp()
	app.Get("/api/v1/invoices", auth.Handler, middleware.RequirePermission("invoices", "read"),
		func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", nil)
	req.Header.Set("Authorization", auth.Bearer(t, tenant))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

// Los permisos se guardan por (tenant, usuario): dos usuarios de la misma
// empresa con permisos distintos nunca comparten la entrada de caché.
func TestRoutes_PermissionsArePerUser_CacheDoesNotMixUsers(t *testing.T) {
	e := newEnv(t, time.Minute)
	e.session.Grant(tenant, "alice", "invoices/manage")
	e.session.Grant(tenant, "bob", "invoices/read")
	alice := e.auth.BearerFor(t, tenant, "alice")
	bob := e.auth.BearerFor(t, tenant, "bob")
	pay := route{http.MethodPatch, "/api/v1/invoices/00000000-0000-4000-8000-000000000001/pay", ""}

	for i := 0; i < 2; i++ {
		status, msg := e.do(t, alice, pay.method, pay.path, "")
		expectAllowed(t, status, msg, pay)
		status, msg = e.do(t, bob, pay.method, pay.path, "")
		expectDenied(t, status, msg, pay)
		if status, msg := e.do(t, bob, http.MethodGet, "/api/v1/invoices", ""); status != http.StatusOK {
			t.Fatalf("bob GET /invoices: expected 200, got %d %q", status, msg)
		}
	}
	if calls := e.session.Calls.Load(); calls != 2 {
		t.Fatalf("expected one session call per user within TTL (2), got %d", calls)
	}

	// Un cambio de permisos se aplica al vencer el TTL; dentro del TTL se usa la caché.
	e.session.Grant(tenant, "bob", "invoices/manage")
	status, msg := e.do(t, bob, pay.method, pay.path, "")
	expectDenied(t, status, msg, pay)
}

func TestRoutes_PermissionChangeAppliesWithoutCache(t *testing.T) {
	e := newEnv(t, 0)
	e.session.Grant(tenant, "bob", "invoices/read")
	bob := e.auth.BearerFor(t, tenant, "bob")
	create := route{http.MethodPost, "/api/v1/customers", `{}`}
	status, msg := e.do(t, bob, create.method, create.path, create.body)
	expectDenied(t, status, msg, create)

	e.session.Grant(tenant, "bob", "customers/manage")
	status, msg = e.do(t, bob, create.method, create.path, create.body)
	expectAllowed(t, status, msg, create)
}

func TestRoutes_SuspendedTenant_StillTenantIsSuspended(t *testing.T) {
	e := newEnv(t, time.Minute)
	e.session.Grant(tenant, "admin", "*/*")
	e.session.Grant(tenant, "other", "*/*")
	e.session.Suspend(tenant, true)

	for _, user := range []string{"admin", "other"} {
		bearer := e.auth.BearerFor(t, tenant, user)
		for _, p := range protected {
			status, msg := e.do(t, bearer, p.method, p.path, p.body)
			if status != http.StatusForbidden || msg != "tenant is suspended" {
				t.Errorf("%s %s %s: expected 403 tenant is suspended, got %d %q", user, p.method, p.path, status, msg)
			}
		}
	}
	// La suspensión se guarda por empresa: el segundo usuario no vuelve a consultar a auth.
	if calls := e.session.Calls.Load(); calls != 1 {
		t.Fatalf("expected 1 session call for a suspended tenant within TTL, got %d", calls)
	}
}
