package wire

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/query"
	"github.com/google/wire"
)

var ApplicationSet = wire.NewSet(
	command.NewAppCreateInvoiceCommand,
	query.NewAppGetByIdInvoiceQuery,
)
