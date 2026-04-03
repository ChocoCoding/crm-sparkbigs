# ══════════════════════════════════════════════════════════════
# CRM SparkBigs — Frontend Angular (Nx) — Multi-stage Dockerfile
# ══════════════════════════════════════════════════════════════

# ── Stage 1: Build Angular ──────────────────────────────────
FROM node:22-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias primero (cache Docker)
COPY package.json package-lock.json ./
RUN npm ci

# Copiar el resto del código fuente
COPY . .

# Build de producción con Nx
RUN npx nx build crm-sparkbigs --configuration=production

# ── Stage 2: Servir con Nginx ───────────────────────────────
FROM nginx:alpine

# Copiar template de nginx (usa $PORT de Railway)
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Copiar los archivos compilados
COPY --from=builder /app/dist/apps/crm-sparkbigs/browser /usr/share/nginx/html

# Railway inyecta PORT automáticamente
ENV PORT=80

EXPOSE ${PORT}

CMD ["nginx", "-g", "daemon off;"]
