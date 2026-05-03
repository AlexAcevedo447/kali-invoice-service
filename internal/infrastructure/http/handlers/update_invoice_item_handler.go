package handlers

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type UpdateInvoiceItemHandler struct {
	updateCmd *appcommand.UpdateInvoiceItemCommand
}

func NewUpdateInvoiceItemHandler(updateCmd *appcommand.UpdateInvoiceItemCommand) *UpdateInvoiceItemHandler {
	return &UpdateInvoiceItemHandler{updateCmd: updateCmd}
}

func (h *UpdateInvoiceItemHandler) Handle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}

	var req appcommand.UpdateInvoiceItemInput
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	item, err := h.updateCmd.Execute(id, req)
	if err != nil {
		logger.Error("failed to update invoice item", logger.Fields{
			"item_id": id,
			"error":   err.Error(),
		})
		return mapDomainError(err)
	}

	metrics.IncrementInvoiceItemsUpdated()
	logger.Info("invoice item updated successfully", logger.Fields{
		"item_id":     item.ID,
		"invoice_id":  item.InvoiceID,
		"unit_price":  item.UnitPrice,
		"quantity":    item.Quantity,
		"total":       item.Total,
		"taxes_count": len(item.Taxes),
	})
	return c.JSON(item)
}
