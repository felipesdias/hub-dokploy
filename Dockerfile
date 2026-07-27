# Stage 1: Build & Security setup
FROM golang:1.23-alpine AS builder

# Install SSL certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

# Create non-root user for maximum container security
RUN adduser -D -g "" -u 10001 appuser

WORKDIR /app

# Cache dependency layer
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Build fully static, stripped, zero-dependency binary (-s -w -trimpath)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -o /tmp/hub-dokploy .

# Stage 2: SCRATCH image (0-byte base image for ultimate lightweight footprint ~7MB total)
FROM scratch

# Copy SSL CA certificates and timezone database
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy user database for non-root execution
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy static Go binary
COPY --from=builder /tmp/hub-dokploy /app/hub-dokploy

# Run as unprivileged non-root user
USER appuser:appuser

# Default runtime environment variables
ENV PORT=8007 \
    DYNAMIC_CONFIG_DIR=/etc/dokploy/traefik/dynamic

EXPOSE 8007

ENTRYPOINT ["/app/hub-dokploy"]
