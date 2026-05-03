//go:build wireinject
// +build wireinject

package di

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/config"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

// InitializeApp es el injector principal declarado para wire.
// wire genera wire_gen.go a partir de este archivo.
func InitializeApp(cfg *config.Config) (*fiber.App, error) {
	wire.Build(
		DatabaseSet,
		RepositorySet,
		ApplicationSet,
		HttpSet,
	)
	return nil, nil
}
