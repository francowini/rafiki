#!/bin/bash

# Deployment script for Hetzner CPX11
# This script should be run on the Hetzner server

set -e

echo "========================================="
echo "Topifier Deployment Script"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "Please run as root or with sudo"
    exit 1
fi

# Get the project root (parent directory of devops)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env"

print_info "Project root: $PROJECT_ROOT"

# Load environment variables from .env file
if [ -f "$ENV_FILE" ]; then
    print_info "Loading environment variables from $ENV_FILE"
    export $(grep -v '^#' "$ENV_FILE" | xargs)
else
    print_warn ".env file not found at $ENV_FILE"
    print_warn "Required: POSTGRES_PASSWORD"
    exit 1
fi

# Change to project root directory for docker compose commands
cd "$PROJECT_ROOT"

# Stop existing containers
print_info "Stopping existing containers..."
docker compose down || true

# Pull latest changes (if using git deployment)
if [ -d .git ]; then
    print_info "Pulling latest changes from git..."
    git pull
fi

# Build and start services
print_info "Building and starting services..."
docker compose up -d --build

# Wait for services to be healthy
print_info "Waiting for services to be healthy..."
sleep 10

# Check service health
print_info "Checking service health..."
if docker compose ps | grep -q "Up"; then
    print_info "Services are running:"
    docker compose ps

    print_info ""
    print_info "========================================="
    print_info "Deployment completed successfully!"
    print_info "========================================="
    print_info "API available at: http://localhost:3000"
    print_info "Debug/Metrics at: http://localhost:3010"
    print_info ""
    print_info "View logs with: docker compose logs -f"
else
    print_error "Service deployment failed!"
    print_error "Check logs with: docker compose logs"
    exit 1
fi
