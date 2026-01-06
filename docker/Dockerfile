# Multi-stage Dockerfile for apprun
# Uses apprun-base for faster incremental builds

# Stage 1: Builder - Compile Go binary
# Use local tag as fallback if GHCR not available
FROM apprun-base:latest AS builder

# Build arguments
ARG BUILD_DATE
ARG VERSION=dev
ARG COMMIT_SHA=unknown

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# Copy source code (dependencies already cached in base image)
COPY core/ ./

# Build static binary with cache optimization
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -p=$(nproc) \
    -o /apprun ./cmd/server/main.go

# Download Atlas CLI for migrations
RUN go install ariga.io/atlas/cmd/atlas@latest && \
    cp /go/bin/atlas /atlas

# Stage 2: Production - Minimal runtime image
FROM alpine:latest

# Use TUNA mirror for faster package downloads
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories

# Install runtime dependencies and create non-root user
RUN apk --no-cache add ca-certificates tzdata wget && \
    adduser -D -u 1000 apprun

WORKDIR /app

# Copy binary, config, and migrations from builder
COPY --from=builder /apprun /app/apprun
COPY --from=builder /atlas /usr/local/bin/atlas
COPY --from=builder /build/config /app/config
COPY --from=builder /build/migrations /app/migrations

# Copy entrypoint script
COPY docker/scripts/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Set ownership
RUN chown -R apprun:apprun /app

# Switch to non-root user
USER apprun

# Expose HTTP and HTTPS ports
EXPOSE 8080 8443

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use entrypoint for migration handling
ENTRYPOINT ["/app/docker-entrypoint.sh"]

# Run application
CMD ["./apprun"]