package messaging

import (
	"context"
	"encoding/json"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/config"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitInvoiceStatusPublisher struct {
	exchange   string
	routingKey string
	channel    *amqp.Channel
}

type noopInvoiceStatusPublisher struct{}

func NewInvoiceStatusPublisher(cfg config.RabbitMQConfig) ports.InvoiceStatusEventPublisher {
	if !cfg.Enabled {
		logger.Info("rabbitmq publisher disabled", logger.Fields{})
		return &noopInvoiceStatusPublisher{}
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		logger.Error("rabbitmq dial failed, using noop publisher", logger.Fields{
			"error": err.Error(),
		})
		return &noopInvoiceStatusPublisher{}
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("rabbitmq channel failed, using noop publisher", logger.Fields{
			"error": err.Error(),
		})
		_ = conn.Close()
		return &noopInvoiceStatusPublisher{}
	}

	if err := ch.ExchangeDeclare(
		cfg.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		logger.Error("rabbitmq exchange declare failed, using noop publisher", logger.Fields{
			"error": err.Error(),
		})
		_ = ch.Close()
		_ = conn.Close()
		return &noopInvoiceStatusPublisher{}
	}

	logger.Info("rabbitmq invoice status publisher enabled", logger.Fields{
		"exchange":    cfg.Exchange,
		"routing_key": cfg.RoutingKey,
	})
	return &rabbitInvoiceStatusPublisher{exchange: cfg.Exchange, routingKey: cfg.RoutingKey, channel: ch}
}

func (p *rabbitInvoiceStatusPublisher) PublishInvoiceStatusChanged(event ports.InvoiceStatusChangedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		logger.Error("failed to marshal invoice status event", logger.Fields{
			"invoice_id": event.InvoiceID,
			"error":      err.Error(),
		})
		metrics.IncrementRabbitEventsFailed()
		return err
	}

	err = p.channel.PublishWithContext(
		context.Background(),
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)

	if err != nil {
		logger.Error("failed to publish invoice status event", logger.Fields{
			"invoice_id":   event.InvoiceID,
			"prev_status":  event.PreviousStatus,
			"new_status":   event.NewStatus,
			"error":        err.Error(),
		})
		metrics.IncrementRabbitEventsFailed()
		return err
	}

	metrics.IncrementRabbitEventsPublished()
	logger.Info("invoice status event published", logger.Fields{
		"invoice_id":   event.InvoiceID,
		"prev_status":  event.PreviousStatus,
		"new_status":   event.NewStatus,
		"exchange":     p.exchange,
		"routing_key":  p.routingKey,
	})

	return nil
}

func (p *noopInvoiceStatusPublisher) PublishInvoiceStatusChanged(event ports.InvoiceStatusChangedEvent) error {
	return nil
}
