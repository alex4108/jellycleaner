FROM golang:1.20-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o jellycleaner ./cmd/jellycleaner

# Create final lightweight image
FROM alpine:latest

# Add ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Create a directory for configuration
RUN mkdir -p /home/appuser/config
VOLUME /home/appuser/config

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder stage
COPY --from=builder /app/jellycleaner .

# Set environment variable for config location
ENV CLEANMEDIA_CONFIG=/home/appuser/config/config.yaml

# Run the application
CMD ["./jellycleaner"]