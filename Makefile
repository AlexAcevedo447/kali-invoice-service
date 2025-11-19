# ============================
# Variables Generales
# ============================

APP_NAME=kali-invoice-service
DEV_COMPOSE=docker compose -f docker-compose.dev.yml
PROD_COMPOSE=docker compose -f docker-compose.prod.yml
DEV_CONTAINER=kali-invoice-dev

GOBIN=$(shell go env GOPATH)/bin
LINTER=$(GOBIN)/golangci-lint

# ============================
# Ayuda
# ============================

help:
	@echo ""
	@echo "===== COMANDOS DISPONIBLES ====="
	@echo " Desarrollo Local:"
	@echo "   make run                - Ejecuta la app (go run)"
	@echo "   make test-local         - Tests locales"
	@echo "   make lint-local         - Linter local"
	@echo "   make fmt                - Formatea código"
	@echo "   make tidy               - go mod tidy"
	@echo "   make tools              - Instala herramientas (golangci-lint)"
	@echo "   make check              - fmt + tidy + lint + test"
	@echo ""
	@echo " Con Docker:"
	@echo "   make dev                - Inicia entorno dev"
	@echo "   make dev-down           - Apaga entorno dev"
	@echo "   make prod               - Inicia entorno prod"
	@echo "   make prod-up            - Prod en segundo plano"
	@echo "   make prod-down          - Apaga prod"
	@echo "   make prod-build         - Build optimizado CI/CD"
	@echo "   make test               - Tests dentro del contenedor"
	@echo "   make lint               - Linter dentro del contenedor"
	@echo "   make lint-fix           - Linter + autofix en contenedor"
	@echo ""

# ============================
# Desarrollo Local
# ============================

run:
	go run ./cmd/api

test-local:
	@echo "Running local tests..."
	go test ./... -cover

fmt:
	go fmt ./...

tidy:
	go mod tidy

tools:
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

lint-local: tools
	@echo "Running golangci-lint locally..."
	@$(LINTER) run

check: fmt tidy lint-local test-local

# ============================
# Desarrollo con Docker
# ============================

dev:
	$(DEV_COMPOSE) up

dev-down:
	$(DEV_COMPOSE) down

# ============================
# Producción (Docker)
# ============================

prod:
	$(PROD_COMPOSE) up

prod-up:
	$(PROD_COMPOSE) up -d

prod-down:
	$(PROD_COMPOSE) down

prod-build:
	docker build --target prod -t $(APP_NAME) .

# ============================
# Tests en Contenedor Dev
# ============================

test:
	docker exec -it $(DEV_CONTAINER) go test ./...

# ============================
# Linter en Contenedor Dev
# ============================

lint:
	docker exec -it $(DEV_CONTAINER) golangci-lint run

lint-fix:
	docker exec -it $(DEV_CONTAINER) golangci-lint run --fix
