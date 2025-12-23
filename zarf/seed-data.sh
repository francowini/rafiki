#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Rafiki Seed Data Script${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEEDDATA_DIR="$SCRIPT_DIR/seeddata"

# Build seed tool if needed
if [ ! -f "$SEEDDATA_DIR/seeddata" ]; then
    echo -e "${YELLOW}Building seed data tool...${NC}"
    (cd "$SEEDDATA_DIR" && go build -o seeddata main.go)
fi

# Load environment variables if .env exists
if [ -f ".env" ]; then
    source .env 2>/dev/null || true
elif [ -f "/opt/rafiki/.env" ]; then
    source /opt/rafiki/.env 2>/dev/null || true
fi

# Validate encryption key
if [ -z "$PARTNER_ENCRYPTION_KEY" ]; then
    echo -e "${RED}Error: PARTNER_ENCRYPTION_KEY not set in .env${NC}"
    echo -e "${YELLOW}Generate key with: openssl rand -hex 32${NC}"
    exit 1
fi

echo -e "${BLUE}Running seed data tool...${NC}"

# Check if using external database
if [ -n "$PARTNER_DB_HOST" ]; then
    DB_URL="postgresql://${PARTNER_DB_USER}:${PARTNER_DB_PASSWORD}@${PARTNER_DB_HOST}:${PARTNER_DB_PORT}/${PARTNER_DB_NAME}?sslmode=${PARTNER_DB_SSLMODE:-require}"
else
    # Use local docker postgres
    POSTGRES_USER=${POSTGRES_USER:-rafiki}
    POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-db}
    POSTGRES_DB=${POSTGRES_DB:-rafiki}
    DB_URL="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable"

    # Check if docker container is running
    if ! docker ps | grep -q rafiki-postgres; then
        echo -e "${RED}Error: rafiki-postgres container not running${NC}"
        echo -e "${YELLOW}Start with: make up${NC}"
        exit 1
    fi
fi

# Temporarily disable errexit to capture exit code
set +e
"$SEEDDATA_DIR/seeddata" "$DB_URL" "$PARTNER_ENCRYPTION_KEY"
RESULT=$?
set -e

if [ $RESULT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}  Seed data created successfully!${NC}"
    echo -e "${GREEN}================================${NC}"
    echo ""
    echo -e "${YELLOW}Test Credentials:${NC}"
    echo "  Email:    admin@rafiki.lat"
    echo "  Password: admin123"
    echo ""
else
    echo -e "${RED}Failed to create seed data${NC}"
    exit 1
fi
