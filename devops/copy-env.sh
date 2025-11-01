#!/bin/bash

# Script to copy .env file to Hetzner server

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Server configuration (hardcoded)
SERVER_IP="178.156.170.37"
SERVER_USER="root"
DEPLOY_PATH="/opt/rafiki"

# Get the project root (parent directory of devops)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env"

print_info "========================================="
print_info "Copy .env to Hetzner Server"
print_info "========================================="

# Check if .env file exists locally
if [ ! -f "$ENV_FILE" ]; then
    print_error ".env file not found at $ENV_FILE!"
    print_info "Please create a .env file in the project root first"
    exit 1
fi

print_info "Found .env file locally"

# Ask for SSH password
print_info "Please enter SSH password for $SERVER_USER@$SERVER_IP:"
read -s SERVER_PASSWORD
echo ""

if [ -z "$SERVER_PASSWORD" ]; then
    print_error "Password cannot be empty!"
    exit 1
fi

# Check if sshpass is installed
if ! command -v sshpass &> /dev/null; then
    print_info "Installing sshpass..."
    brew install hudochenkov/sshpass/sshpass || {
        print_error "Failed to install sshpass"
        print_info "Install manually: brew install hudochenkov/sshpass/sshpass"
        exit 1
    }
fi

# Copy .env file to server
print_info "Copying .env file to server at $DEPLOY_PATH/.env..."
sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no "$ENV_FILE" $SERVER_USER@$SERVER_IP:$DEPLOY_PATH/.env

if [ $? -eq 0 ]; then
    print_info ""
    print_info "========================================="
    print_info ".env file copied successfully!"
    print_info "========================================="
else
    print_error "Failed to copy .env file"
    exit 1
fi
