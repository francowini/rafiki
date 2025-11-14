# Zarf - Development Utilities

This folder contains development scripts and utilities.

## Contents

- **`keys/`** - RSA keys for JWT signing (generated via openssl)
- **`create-user.sh`** - Helper script to create users in local database

---

## Creating Users

### Quick Start

```bash
# Create admin user
./zarf/create-user.sh admin@rafiki.com admin ADMIN "Admin User"

# Create regular user
./zarf/create-user.sh test@example.com password123 USER "Test User"
```

### Usage

```bash
./zarf/create-user.sh <email> <password> [role] [name]
```

**Parameters:**
- `email` - User email (required)
- `password` - User password (required)
- `role` - USER or ADMIN (default: USER)
- `name` - Display name (default: "Test User")

**Examples:**

```bash
# Admin with custom name
./zarf/create-user.sh admin@rafiki.com admin ADMIN "System Admin"

# Regular user
./zarf/create-user.sh john@example.com password123 USER "John Doe"

# User with default role (USER)
./zarf/create-user.sh jane@example.com secret123
```

---

## RSA Keys Setup

### 1. Generate Keys (Local Development)

Generate development RSA keys for JWT signing:

```bash
# Generate 4096-bit RSA key
mkdir -p zarf/keys
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096

# Verify key
openssl rsa -in zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem -check -noout
```

**Key ID (kid):** The filename without `.pem` extension
- Example: `54bb2165-71e1-41a6-af3e-7da4a0e1e2c1`

**IMPORTANT:** Do not commit production keys to Git!

### 2. How Keys are Loaded

**Local Development:**
- Keys in `zarf/keys/` are **copied into Docker image** at build time
- Just run: `docker compose up -d --build`
- Easy and automatic!

**Production:**
- Keys are **mounted from external directory** (not in image)
- Store production keys in `/opt/rafiki/keys` on server
- Use: `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d`
- More secure, keys not in image

See [docs/PRODUCTION_KEYS_SETUP.md](../docs/PRODUCTION_KEYS_SETUP.md) for details.

---

## Manual User Creation

If you prefer to create users manually:

### 1. Generate password hash

```bash
cat > /tmp/gen-hash.go <<'EOF'
package main
import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run gen-hash.go <password>")
        os.Exit(1)
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(hash))
}
EOF

go run /tmp/gen-hash.go "your-password"
```

### 2. Insert into database

```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
```

```sql
INSERT INTO users (
    user_id, name, email, roles, password_hash,
    department, enabled, date_created, date_updated
) VALUES (
    gen_random_uuid(),
    'Your Name',
    'your-email@example.com',
    ARRAY['USER']::TEXT[],
    '$2a$10$YOUR_HASH_HERE',
    NULL,
    true,
    NOW(),
    NOW()
);

-- Verify
SELECT user_id, name, email, roles, enabled FROM users;

\q
```

---

## Common Tasks

### List all users

```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT user_id, name, email, roles, enabled FROM users;"
```

### Delete a user

```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "DELETE FROM users WHERE email='user@example.com';"
```

### Disable a user

```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "UPDATE users SET enabled=false WHERE email='user@example.com';"
```

### Change user role

```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "UPDATE users SET roles=ARRAY['ADMIN']::TEXT[] WHERE email='user@example.com';"
```

---

## Notes

- Users are **only created via SQL** - there is no registration endpoint by design
- Passwords are hashed using bcrypt with default cost (10)
- User deletion cascades to their thinks (foreign key constraint)
- Token expiry is set to 48 hours in the authentication middleware
