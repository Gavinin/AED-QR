# Stage 1: Build Backend (Go)
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app

# Install gcc and musl-dev for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build for current architecture with CGO enabled (targeting musl libc)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s -linkmode external -extldflags '-static'" -tags sqlite_omit_load_extension -o backend ./cmd/backend/main.go

# Stage 2: Build Frontend (Next.js)
FROM node:20-alpine AS frontend-builder
WORKDIR /app

# Copy dependency files
COPY web/package.json web/package-lock.json ./
RUN npm ci

# Copy source code
COPY web/ .

# Build Next.js (Static Export)
ENV NEXT_TELEMETRY_DISABLED 1
RUN npm run build

# Stage 3: Final Runtime Image
FROM node:20-alpine

# Install Nginx, OpenSSL, and runtime dependencies for Go (sqlite needs libc)
RUN apk add --no-cache nginx openssl curl libc6-compat

# Create directory structure
WORKDIR /app
RUN mkdir -p /app/config /app/db /app/log /etc/nginx/ssl /usr/share/nginx/html

# Copy Backend Artifacts
COPY --from=backend-builder /app/backend /app/backend
COPY example /app/example

# Copy Frontend Artifacts (Static Export)
COPY --from=frontend-builder /app/out /usr/share/nginx/html

# Copy Nginx Config
COPY deploy/nginx.conf /etc/nginx/nginx.conf

# Copy Entrypoint Script
COPY deploy/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Environment Variables
ENV NODE_ENV production

# Expose Ports (80 redirect, 443 SSL)
EXPOSE 80 443

# Volumes for persistent data
VOLUME ["/app/config", "/app/db", "/app/log", "/etc/nginx/ssl"]

# Start Everything
ENTRYPOINT ["/entrypoint.sh"]
