# Build stage
FROM golang:1.22-alpine AS build

# Install build dependencies
RUN apk add --no-cache make git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from build stage
COPY --from=build /app/bin/mediate .

# Copy default config file
COPY config.yaml.example /app/config/config.yaml.example

# Create data directory for persistent storage
RUN mkdir -p /app/data
VOLUME /app/data

# Default configuration location
ENV CONFIG_PATH=/app/config/config.yaml

# Run the application
ENTRYPOINT ["./mediate"]
CMD ["--config", "/app/config/config.yaml"]
