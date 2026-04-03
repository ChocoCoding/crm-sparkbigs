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

# Copiar config de nginx para SPA (redirige todo a index.html)
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copiar los archivos compilados
COPY --from=builder /app/dist/apps/crm-sparkbigs/browser /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
