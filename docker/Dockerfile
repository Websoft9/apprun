# Build stage
FROM golang:alpine AS builder

# Set GOPROXY to use multiple CDNs for faster downloads and fallback
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

WORKDIR /app/core

# Copy go mod and sum files first for better caching
COPY core/go.mod core/go.sum ./

# Download dependencies (cached if go.mod unchanged)
RUN go mod download

# Copy source code
COPY core/ ./

# Build the application with parallel compilation
RUN export GOMAXPROCS=$(nproc) && CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Runtime stage
FROM alpine:latest

# Use TUNA mirror for faster package downloads
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/core/main .

# Copy config files
COPY --from=builder /app/core/config ./config

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]