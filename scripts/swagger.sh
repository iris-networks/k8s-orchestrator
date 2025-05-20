#!/bin/sh

# Navigate to the project root
cd "$(dirname "$0")/.."

# Generate swagger documentation
swag init -g internal/api/docs.go -o docs