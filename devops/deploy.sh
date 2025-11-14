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

# Create required directories for Tempo with proper permissions
print_info "Ensuring required directories exist with proper permissions..."
mkdir -p zarf/compose/tempo-data/wal
mkdir -p zarf/compose/tempo-data/blocks
mkdir -p zarf/compose/tempo-data/generator
chmod -R 777 zarf/compose/tempo-data

# Create required directories for Certbot (SSL certificates)
print_info "Ensuring certbot directories exist..."
mkdir -p certbot/www
mkdir -p certbot/conf

# Check for JWT keys (required for authentication)
KEYS_DIR="/opt/rafiki/keys"
if [ ! -d "$KEYS_DIR" ] || [ -z "$(ls -A $KEYS_DIR/*.pem 2>/dev/null)" ]; then
    print_warn "JWT keys not found in $KEYS_DIR"
    print_warn "Authentication will not work without keys!"
    print_info ""
    print_info "To setup production keys, run:"
    print_info "  sudo ./devops/setup-prod-keys.sh"
    print_info ""
    read -p "Continue deployment without keys? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Deployment cancelled. Please setup keys first."
        exit 1
    fi
else
    KEY_COUNT=$(ls -1 $KEYS_DIR/*.pem 2>/dev/null | wc -l | tr -d ' ')
    print_info "Found $KEY_COUNT JWT key(s) in $KEYS_DIR"
fi

# Stop existing containers
print_info "Stopping existing containers..."
docker compose --profile production down || true

# Pull latest changes (if using git deployment)
if [ -d .git ]; then
    print_info "Pulling latest changes from git..."
    git pull
fi

# Build and start services with production profile (includes nginx + certbot)
print_info "Building and starting services with production profile..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile production up -d --build

# Wait for services to be healthy
print_info "Waiting for services to be healthy..."
sleep 10

# Check service health
print_info "Checking service health..."
if docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile production ps | grep -q "Up"; then
    print_info "Services are running:"
    docker compose -f docker-compose.yml -f docker-compose.prod.yml --profile production ps

    print_info ""
    print_info "========================================="
    print_info "Deployment completed successfully!"
    print_info "========================================="
    print_info "API available at: https://api.rafiki.lat (via Nginx)"
    print_info "Direct access: http://localhost:3000 (backend only)"
    print_info "Debug/Metrics at: http://localhost:3010"
    print_info ""

    # Check if keys are loaded
    print_info "Checking authentication status..."
    sleep 3
    if docker compose logs partner-service 2>/dev/null | grep -q "keys_loaded"; then
        KEYS_LOADED=$(docker compose logs partner-service | grep "keys_loaded" | tail -1 | grep -o 'keys_loaded":[0-9]*' | cut -d: -f2)
        print_info "✓ Authentication enabled - Keys loaded: $KEYS_LOADED"
    else
        print_warn "⚠ Could not verify key loading - check logs"
    fi

    print_info ""
    print_info "Next steps:"
    print_info "1. Create admin user: ./zarf/create-user.sh admin@domain.com password ADMIN"
    print_info "2. Ensure DNS is configured for api.rafiki.lat"
    print_info "3. Obtain SSL certificate (see deployment docs)"
    print_info "4. Configure firewall to allow ports 80 and 443"
    print_info ""
    print_info "View logs with: docker compose --profile production logs -f"
else
    print_error "Service deployment failed!"
    print_error "Check logs with: docker compose --profile production logs"
    exit 1
fi
