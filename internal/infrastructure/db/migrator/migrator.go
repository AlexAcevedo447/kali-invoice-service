package migrator

import (
	"context"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/uptrace/bun"
)

// SchemaVersion tracks applied migrations.
type SchemaVersion struct {
	bun.BaseModel `bun:"table:schema_versions,alias:sv"`
	ID            int       `bun:"id,pk,autoincrement"`
	Version       int       `bun:"version,unique"`
	AppliedAt     time.Time `bun:"applied_at"`
}

// Migration represents a database migration.
type Migration struct {
	Version int
	Name    string
	Up      func(context.Context, *bun.DB) error
}

// Migrator manages database schema migrations.
type Migrator struct {
	db *bun.DB
}

// New creates a new Migrator instance.
func New(db *bun.DB) *Migrator {
	return &Migrator{db: db}
}

// Run performs non-versioned creation of tables if they don't exist.
// Kept for backward compatibility. New deployments should use RunVersioned.
func (m *Migrator) Run() error {
	ctx := context.Background()
	tables := []interface{}{
		(*models.InvoiceModel)(nil),
		(*models.InvoiceItemModel)(nil),
		(*models.InvoiceItemTaxModel)(nil),
		(*models.IdempotencyModel)(nil),
	}

	for _, t := range tables {
		if _, err := m.db.NewCreateTable().Model(t).IfNotExists().Exec(ctx); err != nil {
			return err
		}
	}

	logger.Info("non-versioned migrations completed", logger.Fields{})
	return nil
}

// RunVersioned runs all pending versioned migrations.
func (m *Migrator) RunVersioned() error {
	ctx := context.Background()

	// Create schema_versions table if not exists
	if _, err := m.db.NewCreateTable().Model((*SchemaVersion)(nil)).IfNotExists().Exec(ctx); err != nil {
		logger.Error("failed to create schema_versions table", logger.Fields{"error": err.Error()})
		return err
	}

	// Get all migrations
	migrations := m.getMigrations()

	// Get current version
	var currentVersion int
	if err := m.db.NewSelect().Model((*SchemaVersion)(nil)).Order("version DESC").Limit(1).Scan(ctx, &currentVersion); err != nil {
		if err.Error() == "sql: no rows in result set" {
			currentVersion = 0
		} else {
			logger.Error("failed to get current schema version", logger.Fields{"error": err.Error()})
			return err
		}
	}

	// Apply pending migrations
	for _, mig := range migrations {
		if mig.Version <= currentVersion {
			continue
		}

		logger.Info("applying migration", logger.Fields{"version": mig.Version, "name": mig.Name})

		if err := mig.Up(ctx, m.db); err != nil {
			logger.Error("migration failed", logger.Fields{
				"version": mig.Version,
				"name":    mig.Name,
				"error":   err.Error(),
			})
			return err
		}

		// Record migration
		sv := &SchemaVersion{
			Version:   mig.Version,
			AppliedAt: time.Now().UTC(),
		}
		if _, err := m.db.NewInsert().Model(sv).Exec(ctx); err != nil {
			logger.Error("failed to record migration", logger.Fields{
				"version": mig.Version,
				"error":   err.Error(),
			})
			return err
		}

		logger.Info("migration applied successfully", logger.Fields{
			"version": mig.Version,
			"name":    mig.Name,
		})
	}

	return nil
}

// getMigrations returns all available migrations in order.
func (m *Migrator) getMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_invoices_table",
			Up:      m.migrationV001CreateInvoicesTable,
		},
		{
			Version: 2,
			Name:    "create_invoice_items_table",
			Up:      m.migrationV002CreateInvoiceItemsTable,
		},
		{
			Version: 3,
			Name:    "create_invoice_item_taxes_table",
			Up:      m.migrationV003CreateInvoiceItemTaxesTable,
		},
		{
			Version: 4,
			Name:    "create_idempotency_records_table",
			Up:      m.migrationV004CreateIdempotencyTable,
		},
	}
}

// migrationV001CreateInvoicesTable creates the invoices table.
func (m *Migrator) migrationV001CreateInvoicesTable(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().Model((*models.InvoiceModel)(nil)).IfNotExists().Exec(ctx)
	return err
}

// migrationV002CreateInvoiceItemsTable creates the invoice_items table.
func (m *Migrator) migrationV002CreateInvoiceItemsTable(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().Model((*models.InvoiceItemModel)(nil)).IfNotExists().Exec(ctx)
	return err
}

// migrationV003CreateInvoiceItemTaxesTable creates the invoice_item_taxes table.
func (m *Migrator) migrationV003CreateInvoiceItemTaxesTable(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().Model((*models.InvoiceItemTaxModel)(nil)).IfNotExists().Exec(ctx)
	return err
}

// migrationV004CreateIdempotencyTable creates the idempotency_records table.
func (m *Migrator) migrationV004CreateIdempotencyTable(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().Model((*models.IdempotencyModel)(nil)).IfNotExists().Exec(ctx)
	return err
}
