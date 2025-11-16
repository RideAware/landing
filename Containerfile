# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o server ./cmd/landing

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy templates directory
COPY --from=builder /app/templates ./templates

# Copy static files directory
COPY --from=builder /app/static ./static

# Copy .env (optional - can be overridden at runtime)
COPY .env .env

EXPOSE 5000

CMD ["./server"]