# Multi-stage Dockerfile for apprun
# Stage 1: Builder - Compile Go binary
FROM golang:1.24-alpine AS builder

# Set GOPROXY for faster downloads
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go module files
COPY core/go.mod core/go.sum ./
RUN go mod download

# Copy source code
COPY core/ ./

# Build static binary with cache optimization
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -p=$(nproc) \
    -o /apprun ./cmd/server/main.go

# Stage 2: Production - Minimal runtime image
FROM alpine:latest

# Use TUNA mirror for faster package downloads
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories

# Install runtime dependencies and create non-root user
RUN apk --no-cache add ca-certificates tzdata wget && \
    adduser -D -u 1000 apprun

WORKDIR /app

# Copy binary and config from builder
COPY --from=builder /apprun /app/apprun
COPY --from=builder /build/config /app/config

# Set ownership
RUN chown -R apprun:apprun /app

# Switch to non-root user
USER apprun

# Expose HTTP and HTTPS ports
EXPOSE 8080 8443

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run application
CMD ["./apprun"]