package wire

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/commands"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/queries"
	"github.com/google/wire"
)

var DomainSet = wire.NewSet(
	commands.NewCreateInvoiceCommand,
	queries.NewGetInvoiceByIdQuery,
)