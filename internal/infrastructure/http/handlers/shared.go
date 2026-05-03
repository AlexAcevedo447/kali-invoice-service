package handlers

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func validateUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id: must be a valid UUID")
	}
	return nil
}

func mapDomainError(err error) error {
	switch err {
	case invoice.ErrNotFound:
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case invoiceitem.ErrNotFound:
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case invoice.ErrAlreadyPaid, invoice.ErrAlreadyCanceled:
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	default:
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
}

func handleStatusTransition(
	c *fiber.Ctx,
	executeCmd func(string) error,
	getByIDQuery func(string) (*invoice.Invoice, error),
	incrementMetric func(),
	errorMsg string,
	successMsg string,
	retrieveErrorMsg string,
) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}

	if err := executeCmd(id); err != nil {
		logger.Error(errorMsg, logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}

	incrementMetric()
	logger.Info(successMsg, logger.Fields{
		"invoice_id": id,
	})

	inv, err := getByIDQuery(id)
	if err != nil {
		logger.Error(retrieveErrorMsg, logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}
	return c.JSON(inv)
}
