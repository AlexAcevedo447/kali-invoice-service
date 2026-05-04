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
	ensureDatabaseExists(cfg)

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

// ensureDatabaseExists se conecta a "postgres" (BD por defecto) y crea la BD
// de la aplicación si todavía no existe. Solo se usa en entornos locales/dev.
func ensureDatabaseExists(cfg config.DatabaseConfig) {
	adminDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.SSLMode,
	)
	adminDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminDSN)))
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		// Si no podemos conectar al servidor, lo dejará fallar más adelante
		// con el error habitual de conexión.
		return
	}

	var exists bool
	row := adminDB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.DBName,
	)
	if err := row.Scan(&exists); err != nil || exists {
		return
	}

	if _, err := adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.DBName)); err != nil {
		logger.Fatal("failed to create database", logger.Fields{"error": err.Error()})
	}
	logger.Info("database created", logger.Fields{"db": cfg.DBName})
}
