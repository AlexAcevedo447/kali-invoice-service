//go:build wireinject
// +build wireinject

package wire

import (
	netHttp "net/http"

	wireGo "github.com/google/wire"
)

// Inicializa el router con todas las dependencias ya inyectadas
func InitializeInvoiceAPI() netHttp.Handler {
	wireGo.Build(
		DatabaseSet,
		InfrastructureSet,
		ApplicationSet,
		RouterDataSet,
	)
	return nil
}
