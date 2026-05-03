# Kali Invoice Service

Servicio de backend especializado en **gestión de facturas** con arquitectura hexagonal (ports & adapters) y Domain-Driven Design (DDD).

## Responsabilidades de Facturación

✅ **Implementado en este servicio:**
- Creación y actualización de facturas e ítems con impuestos multi-tax
- Impuestos con naturaleza DEBIT (suma) / CREDIT (resta) con monto positivo
- Recálculo transaccional automático de totales al mutar detalles
- Cambios de estado (PENDING → PAID / CANCELED) con transiciones validadas
- Publicación de eventos RabbitMQ al cambiar estado (para sistemas downstream)
- Idempotencia mediante Idempotency-Key con fingerprint y replay 24h
- Migraciones versionadas con tracking en tabla `schema_versions`
- Observabilidad estructurada: logs JSON + métricas de operaciones

❌ **Delegado a servicio externo de seguridad:**
- Autenticación (JWT, OAuth, etc.)
- Autorización (RBAC, ABAC, etc.)
- Resiliencia de mensajes (Outbox pattern)
- Rate limiting
- Encriptación de datos sensibles

---

## Arquitectura

```
├─ internal/
│  ├─ domain/                    # DDD: Entidades, Value Objects, Agregados
│  │  ├─ invoice/               # Agregado de factura
│  │  └─ invoiceitem/           # Agregado de línea con impuestos (DEBIT/CREDIT)
│  │
│  ├─ application/              # Casos de uso: Commands & Queries (CQRS-like)
│  │  ├─ invoice/command/       # CreateInvoice, PayInvoice, CancelInvoice
│  │  ├─ invoice/query/         # GetInvoice, GetAllInvoices
│  │  ├─ invoiceitem/command/   # CreateItem, UpdateItem, DeleteItem
│  │  └─ invoiceitem/query/     # GetItem, GetAllItems
│  │
│  ├─ infrastructure/           # Adapters: DB, HTTP, Messaging, Logger, Config
│  │  ├─ db/                    # Bun ORM + PostgreSQL
│  │  │  ├─ models/            # Models Bun (Invoice, InvoiceItem, InvoiceItemTax)
│  │  │  ├─ repositories/       # Adaptadores de persistencia
│  │  │  ├─ mappers/           # DTO ↔ Model
│  │  │  └─ migrator/          # Migraciones versionadas
│  │  ├─ http/                 # Fiber v2 + Handlers
│  │  ├─ messaging/            # RabbitMQ adapter
│  │  ├─ logger/               # Logs estructurados (JSON)
│  │  ├─ metrics/              # Contadores de operaciones
│  │  └─ config/               # Vars de entorno centralizadas
│  │
│  ├─ ports/                    # Interfaces (contratos de dominio)
│  │  ├─ InvoiceCreator, InvoiceUpdater, ...
│  │  ├─ InvoiceStatusEventPublisher
│  │  └─ IdempotencyRepository
│  │
│  └─ bootstrap/               # Inyección de dependencias (Wire)
│     └─ di/                    # Providers y sets de Wire
│
├─ cmd/
│  └─ api/main.go              # Punto de entrada
│
├─ docker-compose.dev.yml      # Stack local: PostgreSQL + RabbitMQ
├─ docker-compose.prod.yml     # Stack producción
└─ Dockerfile                  # Build multi-stage
```

### Patrones

- **DDD**: Dominio puro sin dependencias externas
- **Hexagonal**: Puertos definen contratos; adapters implementan
- **CQRS-like**: Commands mutate; Queries retrieve
- **Transaccional**: Mutaciones de detalles recalculan totales en misma TX
- **Event-driven**: Estado se publica a RabbitMQ (desacoplado de respuesta HTTP)
- **Idempotencia**: Middleware captura Idempotency-Key, calcula fingerprint, cachea respuesta 24h

---

## Setup

### Requisitos

- Go 1.25.4+
- PostgreSQL 15+
- RabbitMQ 4.0+ (opcional; fallback a noop si DISABLED)
- Docker & Docker Compose (para desarrollo)

### 1. Clonar & Instalar

```bash
git clone https://github.com/AlexAcevedo447/kali-invoice-service.git
cd kali-invoice-service
go mod download
```

### 2. Variables de Entorno

Copiar `.env.example` a `.env`:

```bash
cp .env.example .env
```

Editar `.env` con tu configuración:

```env
# App
APP_PORT=3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=invoices_db
DB_SSLMODE=disable

# RabbitMQ (Opcional)
RABBITMQ_ENABLED=true
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=invoice.events
RABBITMQ_ROUTING_KEY=invoice.status.changed
```

### 3. Iniciar Stack Local (Docker)

```bash
make dev
```

Esto levanta:
- **PostgreSQL** en `localhost:5432`
- **RabbitMQ** admin en `http://localhost:15672` (guest/guest)

### 4. Compilar & Ejecutar

```bash
# Compilar
go build ./...

# Ejecutar
go run ./cmd/api

# O con make
make run
```

Servidor escucha en `http://localhost:3000`

### 5. Tests

```bash
go test ./... -v
```

---

## API Endpoints

### Facturas

#### Crear Factura

```http
POST /api/v1/invoices
Idempotency-Key: unique-key-v1

{
  "customer_id": "cust-123",
  "issue_date": "2026-05-03T10:00:00Z",
  "due_date": "2026-05-10T10:00:00Z",
  "items": [
    {
      "product_description": "Consultoría",
      "quantity": 5,
      "unit_price": 100.00,
      "taxes": [
        {
          "code": "IVA",
          "rate": 19,
          "kind": "DEBIT"          # DEBIT suma, CREDIT resta
        },
        {
          "code": "DESCUENTO",
          "rate": 5,
          "kind": "CREDIT"         # Resta del subtotal
        }
      ]
    }
  ]
}
```

**Response (201 Created):**

```json
{
  "id": "inv-abc123",
  "customer_id": "cust-123",
  "issue_date": "2026-05-03T10:00:00Z",
  "due_date": "2026-05-10T10:00:00Z",
  "status": "PENDING",
  "subtotal": 500.00,
  "tax_total": 65.00,          # IVA: +95, DESCUENTO: -30
  "total": 565.00,
  "items": [
    {
      "id": "item-xyz789",
      "invoice_id": "inv-abc123",
      "product_description": "Consultoría",
      "quantity": 5,
      "unit_price": 100.00,
      "subtotal": 500.00,
      "tax_total": 65.00,       # Cálculo neto por naturaleza
      "total": 565.00,
      "taxes": [
        {
          "code": "IVA",
          "rate": 19,
          "kind": "DEBIT",
          "amount": 95.00
        },
        {
          "code": "DESCUENTO",
          "rate": 5,
          "kind": "CREDIT",
          "amount": 30.00
        }
      ]
    }
  ],
  "created_at": "2026-05-03T10:00:00Z"
}
```

#### Listar Facturas

```http
GET /api/v1/invoices
```

#### Obtener Factura por ID

```http
GET /api/v1/invoices/{id}
```

#### Actualizar Factura

```http
PUT /api/v1/invoices/{id}
Idempotency-Key: unique-key-update-v1

{
  "customer_id": "cust-456",
  "due_date": "2026-05-15T10:00:00Z"
}
```

#### Pagar Factura (Estado: PENDING → PAID)

```http
POST /api/v1/invoices/{id}/pay
Idempotency-Key: unique-key-pay-v1
```

**Publica evento a RabbitMQ:**

```json
{
  "invoice_id": "inv-abc123",
  "previous_status": "PENDING",
  "new_status": "PAID",
  "changed_at": "2026-05-03T10:05:00Z",
  "invoice_snapshot": { ... full invoice JSON ... }
}
```

#### Cancelar Factura (Estado: PENDING/PAID → CANCELED)

```http
POST /api/v1/invoices/{id}/cancel
Idempotency-Key: unique-key-cancel-v1
```

### Ítems de Factura

#### Crear Ítem

```http
POST /api/v1/invoice-items
Idempotency-Key: unique-key-item-v1

{
  "invoice_id": "inv-abc123",
  "product_description": "Desarrollo de API",
  "quantity": 10,
  "unit_price": 200.00,
  "taxes": [
    {
      "code": "IVA",
      "rate": 19,
      "kind": "DEBIT"
    }
  ]
}
```

**Response (201 Created):** Item con totales recalculados

#### Listar Ítems

```http
GET /api/v1/invoice-items
GET /api/v1/invoice-items?invoice_id=inv-abc123   # Filtrar por factura
```

#### Obtener Ítem por ID

```http
GET /api/v1/invoice-items/{id}
```

#### Actualizar Ítem

```http
PUT /api/v1/invoice-items/{id}
Idempotency-Key: unique-key-item-update-v1

{
  "quantity": 15,
  "unit_price": 220.00,
  "taxes": [
    {
      "code": "IVA",
      "rate": 19,
      "kind": "DEBIT"
    },
    {
      "code": "IMPUESTO_CONSUMO",
      "rate": 8,
      "kind": "DEBIT"
    }
  ]
}
```

**Efecto:** 
- Ítem actualizado
- Totales de factura recalculados en transacción (ACID)
- Respuesta retorna ítem + factura actualizada

#### Eliminar Ítem

```http
DELETE /api/v1/invoice-items/{id}
```

**Efecto:**
- Ítem y sus impuestos eliminados
- Totales de factura recalculados
- Response: 204 No Content

### Observabilidad

#### Métricas de Operaciones

```http
GET /api/v1/metrics
```

**Response (200 OK):**

```json
{
  "invoices_created": 42,
  "invoices_paid": 15,
  "invoices_canceled": 3,
  "invoice_items_created": 187,
  "invoice_items_updated": 45,
  "invoice_items_deleted": 12,
  "rabbit_events_published": 18,
  "rabbit_events_failed": 0,
  "idempotency_hits": 5,
  "idempotency_misses": 150
}
```

---

## Impuestos: DEBIT vs CREDIT

### Ejemplo: Factura Colombia

| Concepto | Rate | Kind | Base | Efecto | Monto |
|----------|------|------|------|--------|-------|
| Subtotal | - | - | 500 | - | 500.00 |
| **IVA 19%** | 19 | **DEBIT** | 500 | +95 | **95.00** |
| **Descuento 5%** | 5 | **CREDIT** | 500 | -25 | **-25.00** |
| **Impuesto Consumo 8%** | 8 | **DEBIT** | 500 | +40 | **40.00** |
| **Tax Total** | - | - | - | - | **110.00** |
| **Total** | - | - | - | - | **610.00** |

**Fórmula:**

```
tax_total = sum(amount where kind=DEBIT) - sum(amount where kind=CREDIT)
total = subtotal + tax_total
```

**Almacenamiento:** Amounts siempre positivos; la naturaleza (DEBIT/CREDIT) determina el signo en cálculo.

---

## Flujos de Cambio de Estado

### Estado: Factura Creada

```
CREATE INVOICE
    ↓
Estado = PENDING
    ↓
Totales calculados
    ↓
Response 201
```

### Transición: PENDING → PAID

```
POST /invoices/{id}/pay
    ↓
Valida status = PENDING
    ↓
Status → PAID
    ↓
Publica evento a RabbitMQ
    ↓
Response JSON + evento enviado (fire & forget)
    ↓
Servicio downstream consume evento (ej: auditoría, contabilidad)
```

### Transición: PENDING/PAID → CANCELED

```
POST /invoices/{id}/cancel
    ↓
Valida status ∈ {PENDING, PAID}
    ↓
Status → CANCELED
    ↓
Publica evento a RabbitMQ
    ↓
Response JSON + evento enviado
```

### Validaciones de Transición

- No es posible pasar de CANCELED a otro estado
- Solo desde PENDING o PAID se puede CANCELED
- Eventos incluyen snapshot completo de factura para reproducibilidad

---

## Migraciones Versionadas

### Tabla: `schema_versions`

Registra todas las migraciones aplicadas:

```sql
CREATE TABLE schema_versions (
  id SERIAL PRIMARY KEY,
  version INT UNIQUE NOT NULL,
  applied_at TIMESTAMP NOT NULL
);
```

### Migraciones Disponibles

| Version | Nombre | Descripción |
|---------|--------|-------------|
| 1 | create_invoices_table | Tabla `invoices` con status, totales |
| 2 | create_invoice_items_table | Tabla `invoice_items` con relación 1:N |
| 3 | create_invoice_item_taxes_table | Tabla `invoice_item_taxes` con kind (DEBIT/CREDIT) |
| 4 | create_idempotency_records_table | Tabla `idempotency_records` para replay |

**Ejecución:** Al startup, `migrator.RunVersioned()` aplica solo migrations pendientes.

---

## Observabilidad

### Logs Estructurados (JSON)

Todos los logs están en formato JSON con contexto:

```json
{
  "level": "INFO",
  "message": "invoice created successfully",
  "timestamp": "2026-05-03T10:00:00.123456Z",
  "fields": {
    "invoice_id": "inv-abc123",
    "customer_id": "cust-123",
    "items_count": 2,
    "subtotal": 500.00,
    "tax_total": 65.00,
    "total": 565.00
  }
}
```

### Niveles de Log

- **INFO**: Operaciones exitosas (CREATE, UPDATE, DELETE, PAY, CANCEL)
- **WARN**: Anomalías no fatales (idempotency key reused con payload distinto)
- **ERROR**: Fallos de persistencia, validación, mensajería
- **FATAL**: Errores irrecuperables (startup failure)

### Métrica Global (Endpoint `/api/v1/metrics`)

Acceso de lectura a contadores sin requiere auth en este servicio (delegado a gateway/security).

---

## Integración con Servicio de Seguridad

### Flujo Autenticado (Expected)

```
Client
  │
  ├─→ Security Service (autenticación, JWT)
  │   ├─ Valida token
  │   └─ Propaga usuario en Authorization header
  │
  └─→ Invoice Service (facturación pura)
      ├─ Recibe header Authorization (delegado, no valida aquí)
      ├─ Ejecuta operación
      ├─ Publica evento RabbitMQ si aplica
      └─ Responde al cliente
```

### Rate Limiting, Auditoría, Encriptación

- **Rate Limiting**: Implementar en API Gateway o Security Service
- **Auditoría**: Consumir eventos RabbitMQ en servicio de auditoría (quién, qué, cuándo)
- **Encriptación**: Implementar en API Gateway o middleware de seguridad
- **IP Whitelist**: Configurar en firewall / K8s NetworkPolicy

---

## Variables de Entorno Completa

```env
# === App ===
APP_PORT=3000

# === Database (PostgreSQL) ===
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=invoices_db
DB_SSLMODE=disable

# === RabbitMQ (Optional) ===
RABBITMQ_ENABLED=true                                # false para deshabilitar
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=invoice.events                      # Exchange name
RABBITMQ_ROUTING_KEY=invoice.status.changed           # Routing key para publicar
```

**Notas:**
- Si `RABBITMQ_ENABLED=false`, publica eventos se hace noop (no envía, no rompe)
- Si RabbitMQ no es accesible al startup, fallback a noop publisher
- Las vars se cargan desde `.env` via `github.com/joho/godotenv`

---

## Compilación & Deployment

### Desarrollo Local

```bash
make dev               # Docker Compose con PostgreSQL + RabbitMQ
make run              # go run
make test-local       # go test
make fmt              # Format código
make lint-local       # Lint local
```

### Build Optimizado (CI/CD)

```bash
make prod-build       # Multi-stage Docker build
```

**Dockerfile (Multi-stage):**
1. Stage 1: Go builder (compilar)
2. Stage 2: Alpine runner (runtime mínimo ~15MB)

### Ejecutar en Producción

```bash
docker compose -f docker-compose.prod.yml up -d
```

---

## Validaciones & Restricciones

### Invoice

- `customer_id`: Requerido, string no vacío
- `items[]`: Mínimo 1 ítem
- `status`: Solo PENDING, PAID, CANCELED (enum)
- Transiciones de estado validadas (no cualquier combinación)

### InvoiceItem

- `invoice_id`: Requerido, UUID válido
- `quantity`: > 0
- `unit_price`: ≥ 0
- `taxes[]`: Mínimo 1 impuesto; `kind` ∈ {DEBIT, CREDIT}

### Idempotencia

- `Idempotency-Key`: Requerido en POST/PUT/DELETE
- Fingerprint: SHA256(method|path|body)
- TTL: 24 horas (configurable)
- Reintento con payload distinto: Error 422

---

## Troubleshooting

### "database connection refused"

```bash
docker compose -f docker-compose.dev.yml up -d postgres
# Esperar 10s para que postgres inicie
go run ./cmd/api
```

### "rabbitmq dial failed"

Si ves log: `rabbitmq dial failed, using noop publisher`

Soluciones:
1. Verificar `RABBITMQ_ENABLED=true` en `.env`
2. Verificar RabbitMQ accesible: `amqp-dial $RABBITMQ_URL`
3. Set `RABBITMQ_ENABLED=false` para deshabilitar

### "migration failed"

```bash
# Revisar tabla schema_versions
psql -U postgres -d invoices_db -c "SELECT * FROM schema_versions;"

# Limpiar migraciones (SOLO en DEV!)
psql -U postgres -d invoices_db -c "DROP TABLE IF EXISTS schema_versions, invoices, invoice_items, invoice_item_taxes, idempotency_records CASCADE;"

# Re-ejecutar
go run ./cmd/api
```

---

## Performance & Escalabilidad

### Índices (Automáticos vía Bun)

- Primary keys (id)
- Foreign keys (invoice_id)
- Índice único en `schema_versions.version`
- Índice en `idempotency_records.key` para lookup rápido

### Connection Pooling

- Bun con `maxOpenConns=25`, `maxIdleConns=5`
- RabbitMQ con channel reutilizable

### Transacciones

Todas las mutaciones de detalles (CREATE, UPDATE, DELETE item) recalculan totales en **una sola transacción ACID**.

---

## Roadmap Futuro (No Implementado)

- [ ] Batch operations (crear múltiples facturas en 1 request)
- [ ] Webhooks para eventos (alternativa a RabbitMQ)
- [ ] Soft deletes (facturas archivadas)
- [ ] Attachment support (PDFs, recibos)
- [ ] Multi-currency (COL, USD, EUR, etc.)
- [ ] Payment plan support (cuotas)
- [ ] Approval workflow (DRAFT → SUBMITTED → APPROVED → PAID)
- [ ] Notifications (email, SMS)
- [ ] Export (JSON, CSV, PDF)
- [ ] Advanced search & filtering

---

## Licencia

MIT

---

## Contacto

Para preguntas o reportar bugs: [GitHub Issues](https://github.com/AlexAcevedo447/kali-invoice-service/issues)
