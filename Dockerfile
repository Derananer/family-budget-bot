# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install required build tools and dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev \
    pkgconfig \
    mupdf \
    mupdf-dev \
    build-base

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bankbot

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    tzdata \
    mupdf \
    mupdf-tools

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bankbot .

# Create non-root user
RUN adduser -D appuser
USER appuser

# Command to run the application
CMD ["./bankbot"] 