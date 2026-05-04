# ============================
# Variables Generales
# ============================

APP_NAME=kali-invoice-service
DEV_COMPOSE=docker compose -f docker-compose.dev.yml --env-file .env.dev
PROD_COMPOSE=docker compose -f docker-compose.prod.yml --env-file .env.prod
DEV_CONTAINER=kali-invoice-api-dev
PROD_CONTAINER=kali-invoice-api

GOBIN=$(shell go env GOPATH)/bin
LINTER=$(GOBIN)/golangci-lint

# ============================
# Ayuda
# ============================

help:
	@echo ""
	@echo "===== KALI INVOICE SERVICE - COMANDOS DISPONIBLES ====="
	@echo ""
	@echo "🔧 DESARROLLO LOCAL:"
	@echo "   make run                - Ejecuta la app (go run)"
	@echo "   make test-local         - Tests locales"
	@echo "   make lint-local         - Linter local"
	@echo "   make fmt                - Formatea código"
	@echo "   make tidy               - go mod tidy"
	@echo "   make tools              - Instala herramientas (golangci-lint, wire)"
	@echo "   make check              - fmt + tidy + lint + test (local)"
	@echo "   make wire               - Genera wire_gen.go (dependency injection)"
	@echo ""
	@echo "🐳 DESARROLLO CON DOCKER:"
	@echo "   make dev                - Inicia dev (docker-compose up)"
	@echo "   make dev-logs           - Ver logs en vivo del dev"
	@echo "   make dev-test           - Ejecuta tests en el container dev"
	@echo "   make dev-lint           - Linter en el container dev"
	@echo "   make dev-lint-fix       - Linter + autofix en dev"
	@echo "   make dev-down           - Apaga dev (docker-compose down)"
	@echo "   make dev-clean          - Elimina containers y volúmenes de dev"
	@echo ""
	@echo "🚀 PRODUCCIÓN CON DOCKER:"
	@echo "   make prod               - Inicia prod (docker-compose up -d)"
	@echo "   make prod-logs          - Ver logs en vivo de prod"
	@echo "   make prod-down          - Apaga prod"
	@echo "   make prod-clean         - Elimina containers y volúmenes de prod"
	@echo "   make prod-build         - Build optimizado para CI/CD"
	@echo ""

# ============================
# Desarrollo Local
# ============================

run:
	go run ./cmd/api

test-local:
	@echo "Running local tests..."
	go test ./... -v -cover

fmt:
	go fmt ./...

tidy:
	go mod tidy

tools:
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	@go install github.com/google/wire/cmd/wire@latest

lint-local: tools
	@echo "Running golangci-lint locally..."
	@$(LINTER) run

wire:
	@echo "Generating wire_gen.go..."
	@go run github.com/google/wire/cmd/wire@latest ./internal/bootstrap/di

check: fmt tidy lint-local test-local
	@echo "✅ Local check passed!"

# ============================
# Desarrollo con Docker
# ============================

dev:
	$(DEV_COMPOSE) up --build

dev-logs:
	$(DEV_COMPOSE) logs -f api

dev-test:
	docker exec -it $(DEV_CONTAINER) go test ./... -v

dev-lint:
	docker exec -it $(DEV_CONTAINER) golangci-lint run

dev-lint-fix:
	docker exec -it $(DEV_CONTAINER) golangci-lint run --fix

dev-down:
	$(DEV_COMPOSE) down

dev-clean:
	$(DEV_COMPOSE) down -v --rmi all

# ============================
# Producción (Docker)
# ============================

prod:
	$(PROD_COMPOSE) up -d --build

prod-logs:
	$(PROD_COMPOSE) logs -f api

prod-down:
	$(PROD_COMPOSE) down

prod-clean:
	$(PROD_COMPOSE) down -v --rmi all

prod-build:
	docker build --target prod -t $(APP_NAME):latest -t $(APP_NAME):$$(date +%s) .

# ============================
# Utilidades
# ============================

.PHONY: help run test-local fmt tidy tools check wire lint-local dev dev-logs dev-test dev-lint dev-lint-fix dev-down dev-clean prod prod-logs prod-down prod-clean prod-build
