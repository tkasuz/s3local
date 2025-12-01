# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO is required for sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o s3local ./cmd/s3local

# Runtime stage (production)
FROM alpine:latest AS production

# Install runtime dependencies including Atlas CLI
RUN apk --no-cache add ca-certificates sqlite-libs curl && \
    curl -sSf https://atlasgo.sh | sh

# Create non-root user
RUN addgroup -g 1000 s3local && \
    adduser -D -u 1000 -G s3local s3local

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/s3local .

# Create data directory for database
RUN mkdir -p /data && chown -R s3local:s3local /data

# Switch to non-root user
USER s3local

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV HOST=0.0.0.0
ENV DB_PATH=/data/s3local.db

# Run the application
CMD ["./s3local"]
