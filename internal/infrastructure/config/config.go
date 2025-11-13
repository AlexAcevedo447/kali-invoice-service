package config

import (
	"log"
	"os"
	"strconv"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/database"
	"github.com/joho/godotenv"
)

func LoadDatabaseConfig() *database.PgConfig {
	_ = godotenv.Load()

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Error al convertir DB_PORT: %v", err)
	}

	return &database.PgConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
}