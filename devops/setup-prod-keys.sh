#!/bin/bash

# Script to generate and setup JWT keys on production server
# Run this ONCE on the server after initial deployment

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=========================================${NC}"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "Please run as root or with sudo"
    exit 1
fi

print_header "Production JWT Keys Setup"
echo ""

# Configuration - Keys stored OUTSIDE repo to prevent accidental deletion
KEYS_DIR="/var/lib/rafiki/keys"

# Check if keys already exist
if [ -d "$KEYS_DIR" ] && [ -n "$(ls -A $KEYS_DIR/*.pem 2>/dev/null)" ]; then
    print_warn "Keys already exist in $KEYS_DIR"
    echo ""
    ls -la "$KEYS_DIR"
    echo ""
    read -p "Do you want to generate a NEW key (existing keys will be kept)? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Keeping existing keys. Exiting."
        exit 0
    fi
fi

# Create keys directory
print_info "Creating keys directory: $KEYS_DIR"
mkdir -p "$KEYS_DIR"
chmod 700 "$KEYS_DIR"

# Generate new key ID
NEW_KID=$(uuidgen | tr '[:upper:]' '[:lower:]')
KEY_FILE="$KEYS_DIR/${NEW_KID}.pem"

print_info "Generating new RSA key..."
print_info "Key ID (kid): $NEW_KID"
echo ""

# Generate RSA key
openssl genrsa -out "$KEY_FILE" 4096

if [ $? -eq 0 ]; then
    print_info "Key generated successfully: $KEY_FILE"
else
    print_error "Failed to generate key"
    exit 1
fi

# Set permissions
chmod 600 "$KEY_FILE"
chown root:root "$KEY_FILE"

# Verify key
print_info "Verifying key..."
openssl rsa -in "$KEY_FILE" -check -noout

if [ $? -eq 0 ]; then
    print_info "Key verified successfully"
else
    print_error "Key verification failed"
    exit 1
fi

echo ""
print_header "Keys Directory Contents"
ls -lah "$KEYS_DIR"

echo ""
print_header "Setup Complete!"
echo ""
print_info "Production key details:"
echo "  Directory: $KEYS_DIR"
echo "  Key ID:    $NEW_KID"
echo "  Key File:  $KEY_FILE"
echo ""
print_info "IMPORTANT: Save this Key ID (kid) - you'll need it for authentication!"
echo ""
print_warn "To use this key in login requests:"
echo "  GET /v1/auth/token/${NEW_KID}"
echo ""
print_info "Next steps:"
echo "  1. Run deployment: ./devops/deploy.sh"
echo "  2. Create admin user: ./zarf/create-user.sh admin@domain.com password ADMIN"
echo "  3. Test login with the new kid"
echo ""

# Offer to create backup
read -p "Create encrypted backup of keys? [Y/n] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    BACKUP_FILE="/opt/rafiki/keys-backup-$(date +%Y%m%d-%H%M%S).tar.gz.gpg"
    print_info "Creating encrypted backup..."
    tar -czf - "$KEYS_DIR" | gpg --symmetric --cipher-algo AES256 > "$BACKUP_FILE"

    if [ $? -eq 0 ]; then
        print_info "Backup created: $BACKUP_FILE"
        print_warn "Save the encryption password securely!"
    else
        print_error "Backup creation failed"
    fi
fi

echo ""
print_info "Done! 🎉"
