package routes

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	invoiceRoutes "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/routes/invoice"
	invoiceItemRoutes "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/routes/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// Recursos de permisos de los catálogos del tenant.
const (
	ResourceCustomers = "customers"
	ResourceProducts  = "products"
)

// RegisterDomainRoutes registra rutas agrupadas por dominio de negocio.
func RegisterDomainRoutes(
	router fiber.Router,
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
	catalogH *handlers.CatalogHandler,
	auth fiber.Handler,
) {
	invoiceRoutes.Register(router, createH, updateH, getAllH, getByIDH, payH, cancelH, idempotencyRepo, auth)
	invoiceItemRoutes.Register(router, createItemH, getAllItemH, getItemByIDH, updateItemH, deleteItemH, idempotencyRepo, auth)

	// Catálogos mínimos del tenant: el usuario elige clientes/productos legibles.
	customers := router.Group("/customers", auth)
	customers.Get("/", middleware.RequirePermission(ResourceCustomers, middleware.ActionRead), catalogH.ListCustomers)
	customers.Post("/", middleware.RequirePermission(ResourceCustomers, middleware.ActionManage),
		middleware.Idempotency(idempotencyRepo), catalogH.CreateCustomer)
	products := router.Group("/products", auth)
	products.Get("/", middleware.RequirePermission(ResourceProducts, middleware.ActionRead), catalogH.ListProducts)
	products.Post("/", middleware.RequirePermission(ResourceProducts, middleware.ActionManage),
		middleware.Idempotency(idempotencyRepo), catalogH.CreateProduct)

	// Metrics endpoint for observability (público: sin token ni permisos)
	metricsHandler := handlers.NewMetricsHandler()
	router.Get("/metrics", metricsHandler.Handle)
}
