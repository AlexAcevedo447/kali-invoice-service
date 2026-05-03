package db

import (
	"database/sql"
	"fmt"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/config"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewBunDB(cfg config.DatabaseConfig) *bun.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	bunDB := bun.NewDB(sqldb, pgdialect.New())

	if err := bunDB.Ping(); err != nil {
		logger.Fatal("failed to connect to postgres", logger.Fields{
			"error": err.Error(),
		})
	}

	logger.Info("postgres connection established", logger.Fields{})
	return bunDB
}
