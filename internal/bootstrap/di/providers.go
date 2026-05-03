package di

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	itemcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/command"
	itemquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/config"
	infradb "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/migrator"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/repositories"
	infrahttp "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/routes"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/messaging"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"github.com/uptrace/bun"
)

// --- Extractores y wrappers de providers ---

func provideDatabaseConfig(cfg *config.Config) config.DatabaseConfig {
	return cfg.Database
}

func provideRabbitMQConfig(cfg *config.Config) config.RabbitMQConfig {
	return cfg.RabbitMQ
}

// provideDB crea la conexión y ejecuta las migraciones versionadas antes de devolver el pool.
func provideDB(cfg config.DatabaseConfig) (*bun.DB, error) {
	db := infradb.NewBunDB(cfg)
	if err := migrator.New(db).RunVersioned(); err != nil {
		return nil, err
	}
	return db, nil
}

func provideInvoiceCreateRepo(db *bun.DB) ports.InvoiceCreator {
	return repositories.NewInvoiceCreateRepo(db)
}

func provideInvoiceStatusRepo(db *bun.DB) ports.InvoiceStatusUpdater {
	return repositories.NewInvoiceStatusRepo(db)
}

func provideInvoiceUpdateRepo(db *bun.DB) ports.InvoiceUpdater {
	return repositories.NewInvoiceUpdateRepo(db)
}

func provideInvoiceByIDRepo(db *bun.DB) ports.InvoiceByIDReader {
	return repositories.NewInvoiceByIDRepo(db)
}

func provideInvoiceListRepo(db *bun.DB) ports.InvoiceListReader {
	return repositories.NewInvoiceListRepo(db)
}

func provideInvoiceItemCreateRepo(db *bun.DB) ports.InvoiceItemCreator {
	return repositories.NewInvoiceItemCreateRepo(db)
}

func provideInvoiceItemUpdateRepo(db *bun.DB) ports.InvoiceItemUpdater {
	return repositories.NewInvoiceItemUpdateRepo(db)
}

func provideInvoiceItemDeleteRepo(db *bun.DB) ports.InvoiceItemDeleter {
	return repositories.NewInvoiceItemDeleteRepo(db)
}

func provideInvoiceItemByIDRepo(db *bun.DB) ports.InvoiceItemByIDReader {
	return repositories.NewInvoiceItemByIDRepo(db)
}

func provideInvoiceItemListRepo(db *bun.DB) ports.InvoiceItemListReader {
	return repositories.NewInvoiceItemListRepo(db)
}

func provideIdempotencyRepo(db *bun.DB) ports.IdempotencyRepository {
	return repositories.NewIdempotencyRepo(db)
}

func provideInvoiceStatusEventPublisher(cfg config.RabbitMQConfig) ports.InvoiceStatusEventPublisher {
	return messaging.NewInvoiceStatusPublisher(cfg)
}

func provideApp(
	createH *handlers.CreateInvoiceHandler,
	updateH *handlers.UpdateInvoiceHandler,
	getAllH *handlers.GetAllInvoicesHandler,
	getByIDH *handlers.GetInvoiceByIDHandler,
	payH *handlers.PayInvoiceHandler,
	cancelH *handlers.CancelInvoiceHandler,
	createItemH *handlers.CreateInvoiceItemHandler,
	getAllItemH *handlers.GetAllInvoiceItemsHandler,
	getItemByIDH *handlers.GetInvoiceItemByIDHandler,
	updateItemH *handlers.UpdateInvoiceItemHandler,
	deleteItemH *handlers.DeleteInvoiceItemHandler,
	idempotencyRepo ports.IdempotencyRepository,
) *fiber.App {
	app := infrahttp.NewFiberApp()
	v1 := app.Group("/api/v1")
	routes.RegisterDomainRoutes(
		v1,
		createH,
		updateH,
		getAllH,
		getByIDH,
		payH,
		cancelH,
		createItemH,
		getAllItemH,
		getItemByIDH,
		updateItemH,
		deleteItemH,
		idempotencyRepo,
	)
	return app
}

// --- Provider sets por capa ---

var DatabaseSet = wire.NewSet(
	provideDatabaseConfig,
	provideRabbitMQConfig,
	provideDB,
)

var RepositorySet = wire.NewSet(
	provideInvoiceCreateRepo,
	provideInvoiceStatusRepo,
	provideInvoiceUpdateRepo,
	provideInvoiceByIDRepo,
	provideInvoiceListRepo,
	provideInvoiceItemCreateRepo,
	provideInvoiceItemUpdateRepo,
	provideInvoiceItemDeleteRepo,
	provideInvoiceItemByIDRepo,
	provideInvoiceItemListRepo,
	provideIdempotencyRepo,
	provideInvoiceStatusEventPublisher,
)

var ApplicationSet = wire.NewSet(
	appcommand.NewCreateInvoiceCommand,
	appcommand.NewUpdateInvoiceCommand,
	appcommand.NewPayInvoiceCommand,
	appcommand.NewCancelInvoiceCommand,
	appquery.NewGetAllInvoicesQuery,
	appquery.NewGetInvoiceByIDQuery,
	itemcommand.NewCreateInvoiceItemCommand,
	itemcommand.NewUpdateInvoiceItemCommand,
	itemcommand.NewDeleteInvoiceItemCommand,
	itemquery.NewGetAllInvoiceItemsQuery,
	itemquery.NewGetInvoiceItemByIDQuery,
)

var HttpSet = wire.NewSet(
	handlers.NewCreateInvoiceHandler,
	handlers.NewUpdateInvoiceHandler,
	handlers.NewGetAllInvoicesHandler,
	handlers.NewGetInvoiceByIDHandler,
	handlers.NewPayInvoiceHandler,
	handlers.NewCancelInvoiceHandler,
	handlers.NewCreateInvoiceItemHandler,
	handlers.NewGetAllInvoiceItemsHandler,
	handlers.NewGetInvoiceItemByIDHandler,
	handlers.NewUpdateInvoiceItemHandler,
	handlers.NewDeleteInvoiceItemHandler,
	provideApp,
)
