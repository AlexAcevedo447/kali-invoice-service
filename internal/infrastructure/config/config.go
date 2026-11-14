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
	Auth     AuthConfig
}

// AuthConfig indica cómo validar los access tokens emitidos por kali-auth-context.
type AuthConfig struct {
	JWKSURL  string
	Issuer   string
	Audience string
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

	dbPort, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		log.Fatalf("invalid POSTGRES_PORT: %v", err)
	}

	return &Config{
		AppPort: os.Getenv("APP_PORT"),
		Database: DatabaseConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     dbPort,
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
			SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
		},
		RabbitMQ: RabbitMQConfig{
			Enabled:    os.Getenv("RABBITMQ_ENABLED") == "true",
			URL:        getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:   getEnvOrDefault("RABBITMQ_EXCHANGE", "kali.invoice"),
			RoutingKey: getEnvOrDefault("RABBITMQ_ROUTING_KEY", "invoice.status.changed"),
		},
		Auth: AuthConfig{
			JWKSURL:  mustGetEnv("AUTH_JWKS_URL"),
			Issuer:   os.Getenv("JWT_ISSUER"),
			Audience: os.Getenv("JWT_AUDIENCE"),
		},
	}
}

// mustGetEnv detiene el arranque si falta una variable obligatoria: sin JWKS el
// servicio no podría validar tokens y no debe quedar abierto.
func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
