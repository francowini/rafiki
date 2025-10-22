# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files (create them if they don't exist)
COPY go.* ./
RUN go mod download 2>/dev/null || true

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/partner ./api/services/partner/main.go

# Runtime stage
FROM alpine:3.22

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/partner .

# Expose ports (API and Debug)
EXPOSE 3000 3010

# Run the application
CMD ["./partner"]
