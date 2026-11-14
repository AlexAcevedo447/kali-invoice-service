# ============================
# Variables Generales
# ============================

APP_NAME=kali-invoice-service
DEV_COMPOSE=docker compose -f docker-compose.dev.yml --env-file .env.dev
DEBUG_COMPOSE=docker compose -f docker-compose.debug.yml --env-file .env.dev
PROD_COMPOSE=docker compose -f docker-compose.prod.yml --env-file .env.prod
DEV_CONTAINER=kali-invoice-api-dev
API_PORT?=8080
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
	@echo "   make dev                - Inicia dev en primer plano (docker-compose up)"
	@echo "   make dev-up             - Inicia dev en segundo plano y espera a la API"
	@echo "   make dev-logs           - Ver logs en vivo del dev"
	@echo "   make dev-test           - Ejecuta tests en el container dev"
	@echo "   make dev-lint           - Linter en el container dev"
	@echo "   make dev-lint-fix       - Linter + autofix en dev"
	@echo "   make dev-down           - Apaga dev (docker-compose down)"
	@echo "   make dev-clean          - ⚠️  BORRA containers, imágenes y el VOLUMEN de datos de dev"
	@echo ""
	@echo "🐞 DEBUG:"
	@echo "   make debug              - Inicia stack de debug (Delve en puerto 40000)"
	@echo "   make debug-db           - Levanta solo la BD para debug local"
	@echo "   make debug-logs         - Ver logs del container de debug"
	@echo "   make debug-down         - Apaga debug y restaura dev (conserva volúmenes)"
	@echo "   (Luego conectar desde VS Code: Run & Debug → Go: Docker (Delve remoto))"
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

# Igual que dev, pero en segundo plano y esperando a que la API esté lista.
dev-up:
	$(DEV_COMPOSE) up -d --build
	@$(MAKE) --no-print-directory wait-api

dev-logs:
	$(DEV_COMPOSE) logs -f invoice-api

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
# Debug (Delve remoto)
# ============================

debug:
	@echo "🐞 Quitando contenedores de dev (los volúmenes se conservan) y levantando debug..."
	$(DEV_COMPOSE) stop invoice-api db
	$(DEV_COMPOSE) rm -f invoice-api db
	$(DEBUG_COMPOSE) up --build -d
	@$(MAKE) --no-print-directory wait-api
	@echo "✅ Debug arriba. Delve escuchando en 127.0.0.1:40000 (usa: make debug-logs)"

debug-db:
	@echo "🐞 Levantando solo la BD para debug local..."
	$(DEV_COMPOSE) up db -d

debug-logs:
	$(DEBUG_COMPOSE) logs -f invoice-api

debug-down:
	@echo "🔄 Quitando contenedores de debug (los volúmenes se conservan) y restaurando dev..."
	$(DEBUG_COMPOSE) stop invoice-api db
	$(DEBUG_COMPOSE) rm -f invoice-api db
	$(DEV_COMPOSE) up -d db invoice-api
	@$(MAKE) --no-print-directory wait-api

# Espera a que la API responda en la ruta pública de métricas (no requiere token).
# El listado de facturas exige JWT: respondería 401 aunque la API esté lista.
wait-api:
	@echo "⏳ Esperando API en http://localhost:$(API_PORT)/api/v1/metrics ..."
	@for i in $$(seq 1 90); do \
		code=$$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$(API_PORT)/api/v1/metrics" || true); \
		if [ "$$code" = "200" ]; then \
			echo "✅ API lista en puerto $(API_PORT)"; \
			exit 0; \
		fi; \
		sleep 2; \
	done; \
	echo "❌ Timeout esperando API en puerto $(API_PORT)"; \
	exit 1

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

.PHONY: help run test-local fmt tidy tools check wire lint-local dev dev-up wait-api dev-logs dev-test dev-lint dev-lint-fix dev-down dev-clean debug debug-db debug-logs debug-down prod prod-logs prod-down prod-clean prod-build
