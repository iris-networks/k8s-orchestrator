# syntax=docker/dockerfile:1

# Define the target platform, default to linux/amd64
ARG TARGETPLATFORM=linux/amd64

# --- Builder Stage ---
# Use the build platform to ensure native execution speed and compatibility for build tools.
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

# Install git
RUN apk add --no-cache git

# Copy go module files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Install swag and generate docs. This runs natively on the build machine.
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN /go/bin/swag init

# Cross-compile the application to the TARGETPLATFORM.
ARG TARGETPLATFORM
RUN CGO_ENABLED=0 GOOS=$(echo $TARGETPLATFORM | cut -d'/' -f1) GOARCH=$(echo $TARGETPLATFORM | cut -d'/' -f2) go build -a -o /k8sgo .

# --- Final Stage ---
# Create a minimal production image for the target platform.
FROM --platform=$TARGETPLATFORM alpine:3.18

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the cross-compiled binary and the generated docs from the builder stage
COPY --from=builder /k8sgo .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./k8sgo"]
