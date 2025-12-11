#!/bin/bash
set -e

# Debug mode - uncomment to see all commands
# set -x

DEBUG=${DEBUG:-false}
debug() {
    if [ "$DEBUG" = "true" ]; then
        echo -e "${YELLOW}[DEBUG] $1${NC}"
    fi
}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Rafiki User Creation Script${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Check if arguments provided
if [ $# -eq 0 ]; then
    echo -e "${YELLOW}Usage:${NC}"
    echo "  $0 <email> <password> [role]"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 admin@rafiki.com admin ADMIN"
    echo "  $0 test@example.com password123 USER"
    echo ""
    echo -e "${YELLOW}Roles:${NC} USER, ADMIN (default: USER)"
    exit 1
fi

EMAIL=$1
PASSWORD=$2
ROLE=${3:-USER}
NAME=${4:-"Test User"}

# Validate role
if [[ ! "$ROLE" =~ ^(USER|ADMIN)$ ]]; then
    echo -e "${RED}Error: Role must be USER or ADMIN${NC}"
    exit 1
fi

echo -e "${YELLOW}Creating user:${NC}"
echo "  Email:    $EMAIL"
echo "  Password: $PASSWORD"
echo "  Role:     $ROLE"
echo "  Name:     $NAME"
echo ""

# Generate bcrypt hash using local Go tool
echo -e "${BLUE}Generating password hash...${NC}"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENHASH_DIR="$SCRIPT_DIR/genhash"

debug "SCRIPT_DIR=$SCRIPT_DIR"
debug "GENHASH_DIR=$GENHASH_DIR"

# Build genhash tool if needed
if [ ! -f "$GENHASH_DIR/genhash" ]; then
    echo -e "${YELLOW}Building hash generator...${NC}"
    debug "Building genhash..."
    (cd "$GENHASH_DIR" && go build -o genhash main.go)
    debug "Build complete, exit code: $?"
fi

debug "Running genhash..."
# Generate hash
HASH=$("$GENHASH_DIR/genhash" "$PASSWORD")
debug "Hash result: ${HASH:0:20}..."

if [ -z "$HASH" ]; then
    echo -e "${RED}Failed to generate password hash${NC}"
    exit 1
fi

echo -e "${GREEN}Hash generated successfully${NC}"
echo ""

# Create SQL
SQL="INSERT INTO users (
    user_id,
    name,
    email,
    roles,
    password_hash,
    department,
    enabled,
    date_created,
    date_updated
) VALUES (
    gen_random_uuid(),
    '$NAME',
    '$EMAIL',
    ARRAY['$ROLE']::TEXT[],
    '$HASH',
    NULL,
    true,
    NOW(),
    NOW()
);"

echo -e "${BLUE}Inserting user into database...${NC}"

# Load environment variables if .env exists
debug "Loading .env file..."
if [ -f ".env" ]; then
    debug "Found .env in current directory"
    source .env 2>/dev/null || true
elif [ -f "/opt/rafiki/.env" ]; then
    debug "Found .env in /opt/rafiki"
    source /opt/rafiki/.env 2>/dev/null || true
else
    debug "No .env file found"
fi

debug "PARTNER_DB_HOST=${PARTNER_DB_HOST:-not set}"

# Check if using external database (PlanetScale/Neon/etc)
if [ -n "$PARTNER_DB_HOST" ]; then
    echo -e "${YELLOW}Using external database: $PARTNER_DB_HOST${NC}"

    # Build connection string
    DB_URL="postgresql://${PARTNER_DB_USER}:${PARTNER_DB_PASSWORD}@${PARTNER_DB_HOST}:${PARTNER_DB_PORT}/${PARTNER_DB_NAME}?sslmode=${PARTNER_DB_SSLMODE:-require}"
    debug "DB_URL built (password hidden)"

    # Execute SQL using psql
    debug "Executing psql..."
    psql "$DB_URL" -c "$SQL"
    DB_RESULT=$?
    debug "psql exit code: $DB_RESULT"
else
    # Use local docker postgres
    echo -e "${YELLOW}Using local Docker PostgreSQL${NC}"
    debug "Executing docker exec..."
    debug "SQL: ${SQL:0:100}..."
    docker exec rafiki-postgres psql -U rafiki -d rafiki -c "$SQL"
    DB_RESULT=$?
    debug "docker exec exit code: $DB_RESULT"
fi

if [ $DB_RESULT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ User created successfully!${NC}"
    echo ""
    echo -e "${YELLOW}Login credentials:${NC}"
    echo "  Email:    $EMAIL"
    echo "  Password: $PASSWORD"
    echo ""

    # Show user in database
    echo -e "${BLUE}Verifying in database:${NC}"

    if [ -n "$PARTNER_DB_HOST" ]; then
        psql "$DB_URL" -c "SELECT user_id, name, email, roles, enabled FROM users WHERE email='$EMAIL';"
    else
        docker exec -i rafiki-postgres psql -U rafiki -d rafiki -c \
            "SELECT user_id, name, email, roles, enabled FROM users WHERE email='$EMAIL';"
    fi
else
    echo -e "${RED}Failed to create user${NC}"
    exit 1
fi
