#!/bin/bash

# Diagnostic script to troubleshoot PlanetScale connection issues
# Run this on Hetzner server: sudo ./devops/diagnose-planetscale.sh

set -e

echo "========================================"
echo "PlanetScale Connection Diagnostic"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}  ✓${NC} $1"
}

print_fail() {
    echo -e "${RED}  ✗${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}  !${NC} $1"
}

# Load environment variables
ENV_FILE=".env"
if [ ! -f "$ENV_FILE" ]; then
    print_fail ".env file not found!"
    exit 1
fi

source "$ENV_FILE"

# Test 1: Check if external DB is configured
print_test "Checking if external database is configured..."
if [ -z "$PARTNER_DB_HOST" ]; then
    print_fail "PARTNER_DB_HOST not set in .env"
    echo ""
    echo "Add these to .env:"
    echo "PARTNER_DB_HOST=your-planetscale-host"
    echo "PARTNER_DB_PORT=5432"
    echo "PARTNER_DB_USER=your_user"
    echo "PARTNER_DB_PASSWORD=your_password"
    echo "PARTNER_DB_NAME=your_database"
    echo "PARTNER_DB_SSLMODE=require"
    exit 1
else
    print_pass "PARTNER_DB_HOST=$PARTNER_DB_HOST"
fi

# Test 2: Check all required variables
print_test "Checking required environment variables..."
MISSING=0

for var in PARTNER_DB_HOST PARTNER_DB_PORT PARTNER_DB_USER PARTNER_DB_PASSWORD PARTNER_DB_NAME; do
    if [ -z "${!var}" ]; then
        print_fail "$var is not set"
        MISSING=1
    else
        print_pass "$var is set"
    fi
done

if [ $MISSING -eq 1 ]; then
    exit 1
fi

echo ""

# Test 3: DNS resolution
print_test "Testing DNS resolution for $PARTNER_DB_HOST..."
if nslookup "$PARTNER_DB_HOST" > /dev/null 2>&1; then
    IP=$(nslookup "$PARTNER_DB_HOST" | grep -A1 "Name:" | grep "Address:" | awk '{print $2}' | head -1)
    print_pass "DNS resolves to: $IP"
else
    print_fail "DNS resolution failed"
    exit 1
fi

echo ""

# Test 4: Network connectivity
print_test "Testing network connectivity to $PARTNER_DB_HOST:$PARTNER_DB_PORT..."
if timeout 5 bash -c "cat < /dev/null > /dev/tcp/$PARTNER_DB_HOST/$PARTNER_DB_PORT" 2>/dev/null; then
    print_pass "Port $PARTNER_DB_PORT is reachable"
else
    print_fail "Cannot connect to port $PARTNER_DB_PORT (firewall/network issue?)"
    print_warn "Check if Hetzner IP is whitelisted in PlanetScale"
    exit 1
fi

echo ""

# Test 5: Direct psql connection from host
print_test "Testing direct PostgreSQL connection from host..."
CONNECTION_STRING="postgresql://$PARTNER_DB_USER:$PARTNER_DB_PASSWORD@$PARTNER_DB_HOST:$PARTNER_DB_PORT/$PARTNER_DB_NAME?sslmode=${PARTNER_DB_SSLMODE:-require}"

if command -v psql &> /dev/null; then
    if psql "$CONNECTION_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
        print_pass "Direct psql connection successful"
    else
        print_fail "Direct psql connection failed"
        echo ""
        echo "Try manually:"
        echo "psql \"$CONNECTION_STRING\""
        exit 1
    fi
else
    print_warn "psql not installed, skipping direct connection test"
fi

echo ""

# Test 6: Check Docker container environment
print_test "Checking Docker container environment variables..."
if docker ps --format '{{.Names}}' | grep -q "rafiki-service"; then
    print_pass "Container is running"

    # Check if env vars are passed correctly
    CONTAINER_HOST=$(docker exec rafiki-service printenv PARTNER_DB_HOST 2>/dev/null || echo "NOT_SET")
    CONTAINER_PORT=$(docker exec rafiki-service printenv PARTNER_DB_PORT 2>/dev/null || echo "NOT_SET")

    if [ "$CONTAINER_HOST" = "$PARTNER_DB_HOST" ]; then
        print_pass "PARTNER_DB_HOST passed correctly to container: $CONTAINER_HOST"
    else
        print_fail "PARTNER_DB_HOST mismatch! Host: $PARTNER_DB_HOST, Container: $CONTAINER_HOST"
    fi

    if [ "$CONTAINER_PORT" = "$PARTNER_DB_PORT" ]; then
        print_pass "PARTNER_DB_PORT passed correctly to container: $CONTAINER_PORT"
    else
        print_fail "PARTNER_DB_PORT mismatch! Host: $PARTNER_DB_PORT, Container: $CONTAINER_PORT"
    fi
else
    print_warn "Container not running, skipping container checks"
fi

echo ""

# Test 7: Check container logs for connection errors
print_test "Checking container logs for database errors..."
if docker ps --format '{{.Names}}' | grep -q "rafiki-service"; then
    ERRORS=$(docker logs rafiki-service 2>&1 | grep -i -E "(error|failed|timeout|deadline)" | tail -5)
    if [ -n "$ERRORS" ]; then
        print_fail "Found errors in container logs:"
        echo "$ERRORS"
    else
        print_pass "No obvious errors in logs"
    fi
else
    print_warn "Container not running"
fi

echo ""

# Test 8: Test connection from inside container
print_test "Testing connection from inside Docker container..."
if docker ps --format '{{.Names}}' | grep -q "rafiki-service"; then
    # Try to resolve DNS from inside container
    if docker exec rafiki-service nslookup "$PARTNER_DB_HOST" > /dev/null 2>&1; then
        print_pass "DNS resolution works inside container"
    else
        print_fail "DNS resolution failed inside container"
        print_warn "This might be a Docker DNS issue"
    fi

    # Try to connect to port from inside container
    if docker exec rafiki-service timeout 5 sh -c "cat < /dev/null > /dev/tcp/$PARTNER_DB_HOST/$PARTNER_DB_PORT" 2>/dev/null; then
        print_pass "Network connectivity works inside container"
    else
        print_fail "Cannot connect from inside container"
        print_warn "This is likely a Docker networking issue"
    fi
else
    print_warn "Container not running"
fi

echo ""
echo "========================================"
echo "Diagnostic Summary"
echo "========================================"
echo ""
echo "If all tests passed but app still fails:"
echo "1. Check SSL/TLS certificate verification"
echo "2. Try PARTNER_DB_SSLMODE=disable temporarily (for testing only)"
echo "3. Check PlanetScale connection string format"
echo "4. Verify PlanetScale IP whitelist includes Hetzner server IP"
echo ""
echo "Current Hetzner IP:"
curl -s ifconfig.me
echo ""
echo ""
echo "Deployment command:"
echo "sudo ./devops/deploy.sh"
echo ""
