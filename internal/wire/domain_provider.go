package wire

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/commands"
	"github.com/google/wire"
)

var DomainSet = wire.NewSet(
	wire.Bind(
		new(domain.ICreateInvoiceCommandFactory), 
		new(*commands.CreateInvoiceCommand),
	),
)