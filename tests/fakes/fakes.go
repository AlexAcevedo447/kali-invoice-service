// Package fakes contiene implementaciones en memoria de los puertos de la
// aplicación para pruebas sin base de datos. Replican el contrato de los
// repositorios reales: toda consulta se filtra por tenant y un registro de otro
// tenant se comporta como inexistente.
package fakes

import (
	"sort"
	"strings"
	"sync"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

// ---- Facturas ----

// InvoiceStore implementa InvoiceCreator, InvoiceByIDReader, InvoiceListReader,
// InvoiceUpdater e InvoiceStatusUpdater.
type InvoiceStore struct {
	mu   sync.Mutex
	byID map[string]invoice.Invoice
	// Order conserva el orden de inserción para listar de forma determinista.
	order []string
}

func NewInvoiceStore() *InvoiceStore {
	return &InvoiceStore{byID: map[string]invoice.Invoice{}}
}

func (s *InvoiceStore) Save(inv *invoice.Invoice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[inv.ID] = *inv
	s.order = append(s.order, inv.ID)
	return nil
}

func (s *InvoiceStore) GetByID(tenantID, id string) (*invoice.Invoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.byID[id]
	if !ok || inv.TenantID != tenantID {
		return nil, invoice.ErrNotFound
	}
	return &inv, nil
}

func (s *InvoiceStore) GetAll(tenantID string, p ports.Pagination) ([]*invoice.Invoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var own []*invoice.Invoice
	for _, id := range s.order {
		inv := s.byID[id]
		if inv.TenantID == tenantID {
			own = append(own, &inv)
		}
	}
	return paginate(own, p), nil
}

func (s *InvoiceStore) Count(tenantID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, inv := range s.byID {
		if inv.TenantID == tenantID {
			n++
		}
	}
	return n, nil
}

// Update persiste el encabezado solo si la factura existe dentro de su tenant
// (equivalente al WHERE id = ? AND tenant_id = ? del repositorio real).
func (s *InvoiceStore) Update(inv *invoice.Invoice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.byID[inv.ID]
	if !ok || cur.TenantID != inv.TenantID {
		return invoice.ErrNotFound
	}
	s.byID[inv.ID] = *inv
	return nil
}

func (s *InvoiceStore) UpdateStatus(tenantID, id string, apply func(*invoice.Invoice) error) (*invoice.Invoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.byID[id]
	if !ok || inv.TenantID != tenantID {
		return nil, invoice.ErrNotFound
	}
	if err := apply(&inv); err != nil {
		return nil, err
	}
	s.byID[id] = inv
	return &inv, nil
}

// Raw devuelve la factura sin filtrar por tenant (solo para aserciones).
func (s *InvoiceStore) Raw(id string) (invoice.Invoice, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.byID[id]
	return inv, ok
}

// ---- Publicador de eventos ----

type Publisher struct {
	mu     sync.Mutex
	Events []ports.InvoiceStatusChangedEvent
}

func (p *Publisher) PublishInvoiceStatusChanged(e ports.InvoiceStatusChangedEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Events = append(p.Events, e)
	return nil
}

// ---- Catálogos ----

// CustomerRepo implementa ports.CustomerRepository con unicidad (tenant, documento).
type CustomerRepo struct {
	mu    sync.Mutex
	items []*catalog.Customer
	// CreateErr, si no es nil, se devuelve en Create (para forzar errores).
	CreateErr error
	Creates   int
}

func (r *CustomerRepo) Create(c *catalog.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Creates++
	if r.CreateErr != nil {
		return r.CreateErr
	}
	for _, x := range r.items {
		if x.TenantID == c.TenantID && x.Document == c.Document {
			return catalog.ErrCustomerDocDuplicated
		}
	}
	cp := *c
	r.items = append(r.items, &cp)
	return nil
}

func (r *CustomerRepo) List(tenantID string, q ports.CatalogQuery) ([]*catalog.Customer, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := strings.ToLower(strings.TrimSpace(q.Search))
	var out []*catalog.Customer
	for _, c := range r.items {
		if c.TenantID != tenantID {
			continue
		}
		if s != "" && !strings.Contains(strings.ToLower(c.Name), s) && !strings.Contains(strings.ToLower(c.Document), s) {
			continue
		}
		cp := *c
		out = append(out, &cp)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return paginate(out, q.Pagination), len(out), nil
}

func (r *CustomerRepo) Exists(tenantID, id string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.items {
		if c.ID == id && c.TenantID == tenantID {
			return true, nil
		}
	}
	return false, nil
}

// ProductRepo implementa ports.ProductRepository con unicidad (tenant, código).
type ProductRepo struct {
	mu    sync.Mutex
	items []*catalog.Product
	// CreateErr, si no es nil, se devuelve en Create (para forzar errores).
	CreateErr error
	Creates   int
}

func (r *ProductRepo) Create(p *catalog.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Creates++
	if r.CreateErr != nil {
		return r.CreateErr
	}
	for _, x := range r.items {
		if x.TenantID == p.TenantID && x.Code == p.Code {
			return catalog.ErrProductCodeDuplicated
		}
	}
	cp := *p
	r.items = append(r.items, &cp)
	return nil
}

func (r *ProductRepo) List(tenantID string, q ports.CatalogQuery) ([]*catalog.Product, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := strings.ToLower(strings.TrimSpace(q.Search))
	var out []*catalog.Product
	for _, p := range r.items {
		if p.TenantID != tenantID {
			continue
		}
		if s != "" && !strings.Contains(strings.ToLower(p.Name), s) && !strings.Contains(strings.ToLower(p.Code), s) {
			continue
		}
		cp := *p
		out = append(out, &cp)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return paginate(out, q.Pagination), len(out), nil
}

func (r *ProductRepo) MissingIDs(tenantID string, ids []string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var missing []string
	for _, id := range ids {
		found := false
		for _, p := range r.items {
			if p.ID == id && p.TenantID == tenantID {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

// ---- Idempotencia ----

// IdempotencyRepo implementa ports.IdempotencyRepository en memoria.
type IdempotencyRepo struct {
	mu      sync.Mutex
	Records map[string]*ports.IdempotencyRecord
}

func NewIdempotencyRepo() *IdempotencyRepo {
	return &IdempotencyRepo{Records: map[string]*ports.IdempotencyRecord{}}
}

func (r *IdempotencyRepo) FindByKey(key string) (*ports.IdempotencyRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.Records[key]
	if !ok {
		return nil, nil
	}
	cp := *rec
	return &cp, nil
}

func (r *IdempotencyRepo) Claim(rec *ports.IdempotencyRecord) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.Records[rec.Key]; ok {
		return false, nil
	}
	cp := *rec
	r.Records[rec.Key] = &cp
	return true, nil
}

func (r *IdempotencyRepo) Finalize(key string, statusCode int, body []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rec, ok := r.Records[key]; ok {
		rec.StatusCode = statusCode
		rec.Body = append([]byte(nil), body...)
	}
	return nil
}

// ---- Utilidades ----

func paginate[T any](items []T, p ports.Pagination) []T {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	start := (p.Page - 1) * p.PageSize
	if start >= len(items) {
		return []T{}
	}
	end := start + p.PageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}
