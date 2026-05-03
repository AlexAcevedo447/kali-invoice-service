package main

import (
	"log"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/bootstrap/di"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/config"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
)

func main() {
	cfg := config.Load()

	app, err := di.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	addr := ":" + cfg.AppPort
	logger.Info("server listening on "+addr, logger.Fields{
		"port": cfg.AppPort,
	})
	if err := app.Listen(addr); err != nil {
		logger.Fatal("server error", logger.Fields{
			"error": err.Error(),
		})
	}
}
