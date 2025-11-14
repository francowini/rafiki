#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Copy JWT Keys to Container${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Default values
KEYS_DIR="./zarf/keys"
CONTAINER_NAME="rafiki-service"

# Check if keys directory exists
if [ ! -d "$KEYS_DIR" ]; then
    echo -e "${RED}Error: Keys directory not found: $KEYS_DIR${NC}"
    echo ""
    echo "Generate keys first:"
    echo "  mkdir -p zarf/keys"
    echo "  openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096"
    exit 1
fi

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}Error: Container '$CONTAINER_NAME' is not running${NC}"
    echo ""
    echo "Start services first:"
    echo "  docker compose up -d"
    exit 1
fi

# Count keys
KEY_COUNT=$(find "$KEYS_DIR" -name "*.pem" -type f | wc -l | tr -d ' ')

if [ "$KEY_COUNT" -eq 0 ]; then
    echo -e "${RED}Error: No .pem keys found in $KEYS_DIR${NC}"
    echo ""
    echo "Generate keys first:"
    echo "  openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096"
    exit 1
fi

echo -e "${YELLOW}Found $KEY_COUNT key(s) in $KEYS_DIR${NC}"
echo ""

# List keys to be copied
echo -e "${BLUE}Keys to copy:${NC}"
find "$KEYS_DIR" -name "*.pem" -type f -exec basename {} \; | while read -r key; do
    echo "  - $key"
done
echo ""

# Copy keys to container
echo -e "${BLUE}Copying keys to container...${NC}"

# Create directory in container if it doesn't exist
docker exec "$CONTAINER_NAME" mkdir -p /app/zarf/keys

# Copy each key
find "$KEYS_DIR" -name "*.pem" -type f | while read -r key_file; do
    key_name=$(basename "$key_file")
    docker cp "$key_file" "${CONTAINER_NAME}:/app/zarf/keys/"
    echo -e "  ${GREEN}✓${NC} Copied: $key_name"
done

echo ""
echo -e "${GREEN}✓ Keys copied successfully!${NC}"
echo ""

# Restart the container to load keys
echo -e "${BLUE}Restarting container to load keys...${NC}"
docker compose restart partner-service

echo ""
echo -e "${YELLOW}Waiting for service to start...${NC}"
sleep 3

# Check if service started successfully
if docker compose logs partner-service --tail=20 | grep -q "keys_loaded:"; then
    KEYS_LOADED=$(docker compose logs partner-service --tail=20 | grep "keys_loaded:" | tail -1 | grep -o 'keys_loaded":[0-9]*' | cut -d: -f2)
    echo -e "${GREEN}✓ Service started successfully!${NC}"
    echo -e "  Keys loaded: ${GREEN}$KEYS_LOADED${NC}"
else
    echo -e "${YELLOW}⚠ Service is starting... Check logs:${NC}"
    echo "  docker compose logs -f partner-service"
fi

echo ""
echo -e "${BLUE}You can now test authentication!${NC}"
