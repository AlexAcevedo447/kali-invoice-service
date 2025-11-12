package wire

import (
	netHttp "net/http"

	wireGo "github.com/google/wire"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
)

// Inicializa el router con todas las dependencias ya inyectadas
func InitializeInvoiceAPI() netHttp.Handler {
	wireGo.Build(
		command.NewAppCreateInvoiceCommand,
		query.NewAppGetByIdInvoiceQuery,
		http.NewInvoiceHandler,
		http.NewKaliInvoiceRouter,
	)
	return nil
}
