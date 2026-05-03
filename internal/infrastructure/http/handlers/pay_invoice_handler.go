package handlers

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type PayInvoiceHandler struct {
	payCmd       *appcommand.PayInvoiceCommand
	getByIDQuery *appquery.GetInvoiceByIDQuery
}

func NewPayInvoiceHandler(
	payCmd *appcommand.PayInvoiceCommand,
	getByIDQuery *appquery.GetInvoiceByIDQuery,
) *PayInvoiceHandler {
	return &PayInvoiceHandler{payCmd: payCmd, getByIDQuery: getByIDQuery}
}

func (h *PayInvoiceHandler) Handle(c *fiber.Ctx) error {
	return handleStatusTransition(
		c,
		h.payCmd.Execute,
		h.getByIDQuery.Execute,
		metrics.IncrementInvoicesPaid,
		"failed to pay invoice",
		"invoice paid successfully",
		"failed to retrieve paid invoice",
	)
}
