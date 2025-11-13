package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PgConfig struct {
	Host string
	Port int
	User string
	Password string
	DBName string
	SSLMode string
}

func NewPostgresConnection(cfg *PgConfig) *gorm.DB {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error opening connection %v", err)
	}

	log.Println("Postgres connection successfully established")
	return db
}