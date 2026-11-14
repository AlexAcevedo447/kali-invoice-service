package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

// isUniqueViolation detecta el código SQLSTATE 23505 de Postgres.
func isUniqueViolation(err error) bool {
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) {
		return pgErr.Field('C') == "23505"
	}
	return false
}

func normalizePagination(p ports.Pagination) ports.Pagination {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p
}

// ---- Clientes ----

type customerRepo struct {
	db *bun.DB
}

func NewCustomerRepo(db *bun.DB) ports.CustomerRepository {
	return &customerRepo{db: db}
}

func (r *customerRepo) Create(c *catalog.Customer) error {
	_, err := r.db.NewInsert().Model(mappers.CustomerToModel(c)).Exec(context.Background())
	if err != nil && isUniqueViolation(err) {
		return catalog.ErrCustomerDocDuplicated
	}
	return err
}

func (r *customerRepo) List(tenantID string, q ports.CatalogQuery) ([]*catalog.Customer, int, error) {
	p := normalizePagination(q.Pagination)
	var ms []models.CustomerModel
	sel := r.db.NewSelect().Model(&ms).Where("cu.tenant_id = ?", tenantID)
	if s := strings.TrimSpace(q.Search); s != "" {
		like := "%" + s + "%"
		sel = sel.WhereGroup(" AND ", func(g *bun.SelectQuery) *bun.SelectQuery {
			return g.Where("cu.name ILIKE ?", like).WhereOr("cu.document ILIKE ?", like)
		})
	}
	total, err := sel.OrderExpr("cu.name ASC, cu.id ASC").
		Limit(p.PageSize).
		Offset((p.Page - 1) * p.PageSize).
		ScanAndCount(context.Background())
	if err != nil {
		return nil, 0, err
	}
	out := make([]*catalog.Customer, 0, len(ms))
	for i := range ms {
		out = append(out, mappers.ModelToCustomer(&ms[i]))
	}
	return out, total, nil
}

func (r *customerRepo) Exists(tenantID, id string) (bool, error) {
	return r.db.NewSelect().
		Model((*models.CustomerModel)(nil)).
		Where("id = ?", id).
		Where("tenant_id = ?", tenantID).
		Exists(context.Background())
}

// ---- Productos ----

type productRepo struct {
	db *bun.DB
}

func NewProductRepo(db *bun.DB) ports.ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(p *catalog.Product) error {
	_, err := r.db.NewInsert().Model(mappers.ProductToModel(p)).Exec(context.Background())
	if err != nil && isUniqueViolation(err) {
		return catalog.ErrProductCodeDuplicated
	}
	return err
}

func (r *productRepo) List(tenantID string, q ports.CatalogQuery) ([]*catalog.Product, int, error) {
	pg := normalizePagination(q.Pagination)
	var ms []models.ProductModel
	sel := r.db.NewSelect().Model(&ms).Where("pr.tenant_id = ?", tenantID)
	if s := strings.TrimSpace(q.Search); s != "" {
		like := "%" + s + "%"
		sel = sel.WhereGroup(" AND ", func(g *bun.SelectQuery) *bun.SelectQuery {
			return g.Where("pr.name ILIKE ?", like).WhereOr("pr.code ILIKE ?", like)
		})
	}
	total, err := sel.OrderExpr("pr.name ASC, pr.id ASC").
		Limit(pg.PageSize).
		Offset((pg.Page - 1) * pg.PageSize).
		ScanAndCount(context.Background())
	if err != nil {
		return nil, 0, err
	}
	out := make([]*catalog.Product, 0, len(ms))
	for i := range ms {
		out = append(out, mappers.ModelToProduct(&ms[i]))
	}
	return out, total, nil
}

func (r *productRepo) MissingIDs(tenantID string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var found []string
	err := r.db.NewSelect().
		Model((*models.ProductModel)(nil)).
		Column("id").
		Where("tenant_id = ?", tenantID).
		Where("id IN (?)", bun.In(ids)).
		Scan(context.Background(), &found)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(found))
	for _, id := range found {
		seen[id] = true
	}
	var missing []string
	for _, id := range ids {
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	return missing, nil
}
