package handlers

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type CreateInvoiceItemHandler struct {
	createCmd *appcommand.CreateInvoiceItemCommand
}

func NewCreateInvoiceItemHandler(createCmd *appcommand.CreateInvoiceItemCommand) *CreateInvoiceItemHandler {
	return &CreateInvoiceItemHandler{createCmd: createCmd}
}

func (h *CreateInvoiceItemHandler) Handle(c *fiber.Ctx) error {
	var req appcommand.CreateInvoiceItemInput
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.InvoiceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invoice_id is required")
	}
	if err := validateUUID(req.InvoiceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invoice_id must be a valid UUID")
	}

	if req.ItemID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "item_id is required")
	}
	if err := validateUUID(req.ItemID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "item_id must be a valid UUID")
	}

	item, err := h.createCmd.Execute(req)
	if err != nil {
		logger.Error("failed to create invoice item", logger.Fields{
			"invoice_id": req.InvoiceID,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}

	metrics.IncrementInvoiceItemsCreated()
	logger.Info("invoice item created successfully", logger.Fields{
		"item_id":     item.ID,
		"invoice_id":  req.InvoiceID,
		"unit_price":  item.UnitPrice,
		"quantity":    item.Quantity,
		"total":       item.Total,
		"taxes_count": len(item.Taxes),
	})
	return c.Status(fiber.StatusCreated).JSON(item)
}
