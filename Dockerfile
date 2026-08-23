# ==============================================================================
# Stage 1: Build Stage
# ==============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install system build dependencies required by Prisma engine fetching & SSL
RUN apk add --no-cache git ca-certificates openssl

# Copy module definitions
COPY go.mod go.sum ./
RUN go mod download

# Copy Prisma schema and generate Linux-compatible Prisma Client Go engine
COPY prisma/ ./prisma/
RUN go run github.com/steebchen/prisma-client-go generate --schema=./prisma/schema.prisma

# Copy source code and docs
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/ ./docs/

# Build statically-linked Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/server \
    ./cmd/server/main.go

# ==============================================================================
# Stage 2: Minimal Runtime Stage
# ==============================================================================
FROM alpine:3.21 AS runner

# Install runtime dependencies for Prisma query engine and TLS certificates
RUN apk add --no-cache ca-certificates openssl tzdata curl

# Create unprivileged application user for security
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

WORKDIR /app

# Copy compiled binary and API documentation
COPY --from=builder --chown=appuser:appgroup /app/bin/server /app/server
COPY --from=builder --chown=appuser:appgroup /app/docs /app/docs

# Switch to unprivileged user
USER appuser:appgroup

# Default environment configuration
ENV SERVER_PORT=8082 \
    ENV=production

EXPOSE 8082

# Container health probe
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1

ENTRYPOINT ["/app/server"]
