# Local Testing Steps

Quick guide to test the authentication system locally.

## Prerequisites

1. **Generate RSA Key** (one-time setup):
```bash
mkdir -p zarf/keys
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096
```

2. **Install Bruno** (if not installed):
```bash
brew install bruno
```

---

## Step 1: Start Services

```bash
# Clean start
docker compose down -v

# Start services (keys from zarf/keys/ are copied automatically)
docker compose up -d --build

# Watch logs
docker compose logs -f partner-service

# Wait for:
# - "keys_loaded: 1"
# - "authentication support enabled"
# - "api router started"
```

**Note:** In local dev, keys from `zarf/keys/` are included in the Docker image automatically.

---

## Step 2: Create Test User

### Generate password hash:
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

# Generate hash for "password123"
go run /tmp/gen-hash.go "password123"
```

### Insert user:
```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
```

```sql
INSERT INTO users (
    user_id, name, email, roles, password_hash,
    department, enabled, date_created, date_updated
) VALUES (
    gen_random_uuid(),
    'Test User',
    'test@example.com',
    ARRAY['USER']::TEXT[],
    'PASTE_HASH_HERE',
    NULL,
    true,
    NOW(),
    NOW()
);

-- Verify
SELECT user_id, name, email, enabled FROM users;

\q
```

---

## Step 3: Test with Bruno

### 3.1 Open Bruno
1. Open Bruno app
2. File → Open Collection
3. Select `/Users/francowini/Documents/rafiki/bruno/rafiki`
4. Select environment: **"local"**

### 3.2 Test Health (no auth needed)
1. Open **"Partner local readiness"**
2. Click **Send**
3. Should get: `200 OK` with `{"status":"ok"}`

### 3.3 Login to Get Token
1. Open **"Login - Get Token"**
2. Click **Send**
3. Response contains JWT token
4. **The token is automatically saved to `jwt-token` variable**

### 3.4 Create Think (authenticated)
1. Open **"Create Think (Authenticated)"**
2. Click **Send**
3. Should get: `201 Created` with think object

### 3.5 Get All Thinks (authenticated)
1. Open **"Get all thinks (Authenticated)"**
2. Click **Send**
3. Should get: `200 OK` with array of your thinks

### 3.6 Test Without Token (should fail)
1. Open **"Create Think (NO AUTH - Should Fail)"**
2. Click **Send**
3. Should get: `401 Unauthorized`

---

## Step 4: Test User Isolation

### Create second user:
```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
```

```sql
-- Generate new hash first: go run /tmp/gen-hash.go "password456"

INSERT INTO users (
    user_id, name, email, roles, password_hash,
    department, enabled, date_created, date_updated
) VALUES (
    gen_random_uuid(),
    'User Two',
    'user2@example.com',
    ARRAY['USER']::TEXT[],
    'PASTE_NEW_HASH_HERE',
    NULL,
    true,
    NOW(),
    NOW()
);

\q
```

### Test in Bruno:
1. Open **"Login - Get Token"**
2. Change email to `user2@example.com` and password to `password456`
3. Click **Send** (token auto-saved)
4. Open **"Get all thinks (Authenticated)"**
5. Click **Send**
6. Should get empty array (user2 has no thinks)

---

## Verification Checklist

- [ ] Service starts without errors
- [ ] Health endpoints work (no auth)
- [ ] Login returns JWT token
- [ ] Token auto-saves to variable
- [ ] Can create think with token
- [ ] Can query thinks with token
- [ ] Cannot create without token (401)
- [ ] User2 cannot see User1's thinks

---

## Troubleshooting

### "key not found" error
```bash
# Verify key exists
ls -la zarf/keys/

# Regenerate if needed
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096
```

### Login returns 401
```bash
# Check user exists and enabled
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT email, enabled FROM users WHERE email='test@example.com';"

# Verify password hash starts with $2a$ or $2b$
```

### Token not working
```bash
# Check service logs
docker compose logs partner-service | grep -i error
docker compose logs partner-service | grep -i opa
```

---

## Next Steps

After local testing works:
1. Deploy to production (see deployment docs)
2. Create more test users as needed
3. Integrate with frontend
