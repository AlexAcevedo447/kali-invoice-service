# ====================================
# Etapa base
# ====================================
FROM golang:1.25.4-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# ====================================
# Etapa desarrollo
# ====================================
FROM base AS dev
RUN apk add --no-cache git bash build-base wget

# Hot reload
RUN go install github.com/air-verse/air@latest

# Linter oficial
RUN wget -O- -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b /usr/local/bin v1.59.1

COPY . .

# Descargar dependencias del proyecto
RUN go mod download

CMD ["air", "-c", ".air.toml"]

# ====================================
# Etapa build (producción)
# ====================================
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /kali-invoice ./cmd/api

# ====================================
# Etapa producción
# ====================================
FROM alpine:3.19 AS prod
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /kali-invoice .
EXPOSE 3000
CMD ["./kali-invoice"]
