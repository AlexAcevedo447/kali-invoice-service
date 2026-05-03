package handlers

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	snapshot := metrics.GetSnapshot()
	return c.JSON(snapshot)
}
