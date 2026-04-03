# ══════════════════════════════════════════════════════════════
# CRM SparkBigs — Multi-stage Dockerfile (Backend Go + Fiber)
# ══════════════════════════════════════════════════════════════

# ── Stage 1: Build ──────────────────────────────────────────
FROM golang:1-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copiar dependencias primero (cache de capas Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY . .

# Compilar binario estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /server ./cmd/server

# ── Stage 2: Runtime ────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copiar solo el binario compilado
COPY --from=builder /server .

# Crear directorio para uploads
RUN mkdir -p /app/uploads

# Puerto por defecto (Railway inyecta PORT automáticamente)
EXPOSE 8080

# Ejecutar
CMD ["./server"]
