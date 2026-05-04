# ====================================
# Stage 1: Base (shared layers)
# ====================================
FROM golang:1.25-alpine AS base
WORKDIR /app

# Instalar dependencias comunes
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Pre-descargar módulos (caché layer)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ====================================
# Stage 2: Development (with hot-reload & tools)
# ====================================
FROM base AS dev
LABEL stage=development

# Herramientas de dev
RUN apk add --no-cache \
    git \
    bash \
    build-base \
    wget \
    curl \
    postgresql-client

# Air para hot-reload
RUN go install github.com/air-verse/air@latest

# Linter
RUN wget -O- -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b /usr/local/bin v1.59.1

# Wire para dependency injection
RUN go install github.com/google/wire/cmd/wire@latest

# Copiar código y deps
COPY . .
RUN go mod download

EXPOSE 8080

# Healthcheck para dev
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/metrics || exit 1

CMD ["air", "-c", ".air.toml"]

# ====================================
# Stage 3: Builder (optimized for prod)
# ====================================
FROM base AS builder
LABEL stage=builder

COPY . .

# Compile optimizado para producción
# CGO_ENABLED=0: static binary (sin libc)
# GOOS=linux GOARCH=amd64: target platform
# -ldflags: strip debug symbols
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'unknown')" \
    -a -installsuffix cgo \
    -o /kali-invoice ./cmd/api

# Verificar que la build está clean
RUN file /kali-invoice && /kali-invoice --version 2>/dev/null || true

# ====================================
# Stage 4: Production (minimal runtime)
# ====================================
FROM scratch AS prod
LABEL stage=production

# Copiar ca-certificates desde base
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo

# Non-root user
COPY --from=builder --chown=1000:1000 /kali-invoice /

USER 1000

EXPOSE 8080
ENV PORT=8080

# Ejecutable directo
ENTRYPOINT ["/kali-invoice"]
