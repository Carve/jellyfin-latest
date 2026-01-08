# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy source code
COPY main.go ./

# Build the application
RUN go build -ldflags="-w -s" -o jellyfin-latest .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -S jellyfin-latest && adduser -S jellyfin-latest -G jellyfin-latest

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/jellyfin-latest .

# Change ownership
RUN chown -R jellyfin-latest:jellyfin-latest /app

# Switch to non-root user
USER jellyfin-latest

# Expose port
EXPOSE 7654

# Run the application
CMD ["./jellyfin-latest"]
