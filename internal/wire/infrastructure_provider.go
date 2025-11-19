package wire

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/contracts"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/postgres"
	"github.com/google/wire"
)

var InfrastructureSet = wire.NewSet(
		// Commands
		postgres.NewInvoiceCommandRepository,
		wire.Bind(new(contracts.ICreateInvoiceCommandRepository), new(*postgres.InvoiceCommandRepository)),
	
		// Queries
		postgres.NewInvoiceQueryRepository,
		wire.Bind(new(contracts.IGetByIdInvoiceQueryRepository), new(*postgres.InvoiceQueryRepository)),
		wire.Bind(new(contracts.IGetAllInvoiceQueryRepository), new(*postgres.InvoiceQueryRepository)),
)
