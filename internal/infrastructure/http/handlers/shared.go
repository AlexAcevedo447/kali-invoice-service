package handlers

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
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