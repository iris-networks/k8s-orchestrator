FROM --platform=linux/amd64 golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install required packages
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies (if go.sum exists)
RUN if [ -f go.sum ]; then go mod download; else go mod tidy; fi

# Copy source code
COPY . .

# Generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init

# Build the application with explicit GOOS and GOARCH for Linux AMD64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o k8sgo .

# Create a minimal production image
FROM --platform=linux/amd64 alpine:3.18

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/k8sgo .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./k8sgo"]