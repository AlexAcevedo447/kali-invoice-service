package wire

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/config"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/database"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
	"github.com/google/wire"
)

var DatabaseSet = wire.NewSet(
	config.LoadDatabaseConfig,
	database.NewPostgresConnection,
)

func NewRouterHandlers(
	invoiceHandler *http.InvoiceHandler,
) []http.RouterHandler {
	return []http.RouterHandler{
		invoiceHandler,
	}
}

var RouterDataSet = wire.NewSet(
	http.NewInvoiceHandler,
	NewRouterHandlers,
	http.NewKaliInvoiceRouter,
)

