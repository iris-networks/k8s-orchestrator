FROM golang:1.21 as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger docs
RUN mkdir -p docs
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12
RUN swag init -g internal/api/docs.go -o docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o k8sgo main.go

# Use distroless as minimal base image
FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=builder /app/k8sgo /k8sgo
COPY --from=builder /app/docs /docs

EXPOSE 8080

ENTRYPOINT ["/k8sgo"]