# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Set Go proxy
ENV GOPROXY=https://goproxy.cn,direct

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gpm cmd/gpm/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary and default config
COPY --from=builder /app/gpm .
COPY --from=builder /app/configs/config.yaml ./configs/config.yaml

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./gpm"]
CMD ["-config", "configs/config.yaml"]
