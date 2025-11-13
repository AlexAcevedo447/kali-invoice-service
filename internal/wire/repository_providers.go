package wire

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/postgres"
	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	// Commands
	postgres.NewInvoiceCommandRepository,
	wire.Bind(new(domain.ICreateInvoiceCommandRepository), new(*postgres.InvoiceCommandRepository)),

	// Queries
	postgres.NewInvoiceQueryRepository,
	wire.Bind(new(domain.IGetByIdInvoiceQueryRepository), new(*postgres.InvoiceQueryRepository)),
	wire.Bind(new(domain.IGetAllInvoiceQueryRepository), new(*postgres.InvoiceQueryRepository)),
)
