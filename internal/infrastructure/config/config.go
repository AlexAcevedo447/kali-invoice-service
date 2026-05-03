package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort  string
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RabbitMQConfig struct {
	Enabled    bool
	URL        string
	Exchange   string
	RoutingKey string
}

func Load() *Config {
	_ = godotenv.Load()

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("invalid DB_PORT: %v", err)
	}

	return &Config{
		AppPort: os.Getenv("APP_PORT"),
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     dbPort,
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
		RabbitMQ: RabbitMQConfig{
			Enabled:    os.Getenv("RABBITMQ_ENABLED") == "true",
			URL:        getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:   getEnvOrDefault("RABBITMQ_EXCHANGE", "kali.invoice"),
			RoutingKey: getEnvOrDefault("RABBITMQ_ROUTING_KEY", "invoice.status.changed"),
		},
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
