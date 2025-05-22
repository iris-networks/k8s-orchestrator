# Kubernetes Sandbox Platform

A Golang service that manages per-user containerized sandboxes using Kubernetes APIs.

## Overview

This service allows for creating and managing isolated user environments (sandboxes) in a Kubernetes cluster. Each sandbox provides:

- Linux desktop environment with Firefox browser
- VNC access via web browser
- Persistent storage that survives container restarts
- Unique subdomains for each user

## Features

- **Container Management**: Create/delete user sandboxes using `accetto/ubuntu-vnc-xfce-firefox-g3` image
- **Persistent Storage**: Attach user-specific persistent volumes
- **Dynamic Subdomains**: Provision unique subdomains per user using Traefik
- **REST API**: Endpoints for container lifecycle management

## Prerequisites

- Kubernetes cluster with Traefik ingress controller
- Docker
- Go 1.21+
- `kubectl` configured with appropriate permissions

## Installation

### Clone the repository

```bash
git clone https://github.com/shanurcsenitap/irisk8s.git
cd irisk8s
```

### Install dependencies

```bash
make deps
```

### Build the application

```bash
make build
```

### Generate Swagger documentation

```bash
make swagger
```

## Usage

### Build and Push Docker image

```bash
make docker-all
```

This will:
1. Generate Swagger documentation
2. Build the Docker image
3. Push the image to Docker Hub at `shanurcsenitap/irisk8s:latest`

### Run locally

```bash
make run
```

The API will be available at http://localhost:8080

Swagger documentation will be available at http://localhost:8080/swagger/index.html

## API Endpoints

- `POST /v1/sandbox/{userId}` - Create user sandbox
- `DELETE /v1/sandbox/{userId}` - Delete user sandbox

## Deployment

For deployment instructions to Google Kubernetes Engine, see the [GKE Deployment Guide](docs/deployment/gke-deployment-guide.md).

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.