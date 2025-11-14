# Authentication Deployment - Quick Reference Guide

**Created:** 2025-11-12
**Status:** Ready for Implementation
**Implementation Time:** 1-2 days

---

## Overview

This is the quick reference for deploying JWT authentication to Rafiki. For detailed instructions, see:
- Backend: [BACKEND_AUTH_IMPLEMENTATION_PLAN.md](./BACKEND_AUTH_IMPLEMENTATION_PLAN.md)
- Frontend: [FRONTEND_AUTH_IMPLEMENTATION_PLAN.md](./FRONTEND_AUTH_IMPLEMENTATION_PLAN.md)

---

## Critical Numbers

- **Backend Bugs to Fix:** 11
- **Frontend Files to Create:** 3
- **Frontend Files to Modify:** 5
- **Token Expiry:** 48 hours
- **User Creation:** SQL only (no registration endpoint)
- **Key ID:** `54bb2165-71e1-41a6-af3e-7da4a0e1e2c1`

---

## Implementation Sequence

```
Day 1 Morning: Fix Backend (4-6 hours)
    ↓
Day 1 Afternoon: Test Locally + Deploy Hetzner (2-3 hours)
    ↓
Day 2 Morning: Implement Frontend (4-6 hours)
    ↓
Day 2 Afternoon: Deploy Frontend + E2E Testing (2-3 hours)
```

---

## Backend - 11 Critical Bugs

### Quick Fix List

1. **migrate.sql:29** - Add comma after PRIMARY KEY
2. **thinkbus/model.go** - Add UserID to Think struct
3. **thinkbus/model.go** - Add UserID to NewThink struct
4. **thinkdb/model.go** - Add UserID to database model
5. **thinkdb/thinkdb.go** - Add user_id to CREATE query
6. **thinkdb/thinkdb.go** - Add user_id filter to Query method
7. **thinkdb/thinkdb.go** - Add user_id filter to Count method
8. **thinkdb/thinkdb.go** - Add user_id to QueryByID method
9. **main.go:168** - Initialize Auth (KeyStore + UserBus + Auth)
10. **mux.go** - Add UserBus and Auth to BusConfig
11. **all.go** - Call authapp.Routes()
12. **thinkapp/route.go** - Add Bearer middleware
13. **thinkapp/thinkapp.go** - Extract user_id with mid.GetSubjectID()
14. **authen.go:73** - Change token expiry to 48 hours

---

## Generate RSA Keys

### Development (Commit to repo)
```bash
mkdir -p zarf/keys
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096
chmod 600 zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
git add -f zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
```

### Production (On Hetzner, don't commit)
```bash
ssh root@178.156.170.37
cd /opt/rafiki
mkdir -p zarf/keys
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096
chmod 600 zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
cp zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem ~/rafiki-prod-key-backup.pem
```

---

## Create Test User (SQL)

### Generate Password Hash
```bash
cat > /tmp/gen-hash.go <<'EOF'
package main
import ("fmt"; "os"; "golang.org/x/crypto/bcrypt")
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
EOF

go run /tmp/gen-hash.go "password123"
```

### Insert User
```sql
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

INSERT INTO users (user_id, name, email, roles, password_hash, department, enabled, date_created, date_updated)
VALUES (
    gen_random_uuid(),
    'Test User',
    'test@example.com',
    ARRAY['USER']::TEXT[],
    '$2a$10$YOUR_HASH_HERE',
    NULL,
    true,
    NOW(),
    NOW()
);

SELECT user_id, email, roles FROM users;
\q
```

---

## Test Backend Locally

```bash
# Start services
docker compose up -d --build
docker compose logs -f partner-service

# Test health
curl http://localhost:3000/v1/readiness

# Test login
curl -X GET \
  -H "Authorization: Basic $(echo -n 'test@example.com:password123' | base64)" \
  http://localhost:3000/v1/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# Save token
TOKEN="paste-token-here"

# Test authenticated think creation
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"category":"personal","content":"test"}' \
  http://localhost:3000/v1/thinks

# Test query thinks
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/thinks
```

---

## Deploy to Hetzner

```bash
# SSH to server
ssh root@178.156.170.37
cd /opt/rafiki

# Generate production key (if not done)
mkdir -p zarf/keys
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096
chmod 600 zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem

# Wipe database and redeploy
docker compose --profile production down -v
git pull origin main
./devops/deploy.sh

# Create admin user (use SQL above)
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Test production
curl http://localhost:3000/v1/readiness
```

---

## Frontend - Files to Create

### 1. Auth Context
**Path:** `frontend/lib/auth-context.tsx`
- Manages auth state (user, token, isAuthenticated)
- Provides login() and logout() methods
- Persists to localStorage

### 2. Login Page
**Path:** `frontend/app/login/page.tsx`
- Email/password form
- React Hook Form + Zod validation
- Error handling

### 3. Protected Route
**Path:** `frontend/components/auth/ProtectedRoute.tsx`
- Redirects to /login if not authenticated
- Shows loading state during auth check

---

## Frontend - Files to Modify

### 4. Root Layout
**Path:** `frontend/app/layout.tsx`
- Wrap with `<AuthProvider>`

### 5. Header
**Path:** `frontend/components/layout/Header.tsx`
- Show user email and logout button
- Show login button when not authenticated

### 6. Thinks Page
**Path:** `frontend/app/thinks/page.tsx`
- Wrap with `<ProtectedRoute>`

### 7. API Client
**Path:** `frontend/lib/api.ts`
- Inject Bearer token from localStorage
- Handle 401 → logout and redirect

### 8. Types
**Path:** `frontend/lib/types.ts`
- Add User, AuthTokenResponse, DecodedToken types

---

## Test Frontend

```bash
cd frontend
npm install
npm run dev

# Open http://localhost:3000/login
# Enter: test@example.com / password123
# Should redirect to /thinks

# Check localStorage
localStorage.getItem('rafiki_auth_token')
localStorage.getItem('rafiki_user')

# Test logout
# Click logout button
# Should redirect to /login
```

---

## Deploy Frontend to Vercel

```bash
# Set environment variable in Vercel Dashboard
NEXT_PUBLIC_API_URL=https://api.rafiki.lat

# Deploy
git add .
git commit -m "feat: Implement login authentication"
git push origin main

# Vercel auto-deploys
```

---

## End-to-End Test

1. ✅ Visit https://rafiki.vercel.app/login
2. ✅ Login with production credentials
3. ✅ Create a think
4. ✅ Logout
5. ✅ Verify cannot access /thinks without login
6. ✅ Login again
7. ✅ Verify only see own thinks

---

## Key Files Reference

### Backend
```
business/sdk/migrate/sql/migrate.sql          # Migration SQL
business/domain/thinkbus/model.go             # Think business model
business/domain/thinkbus/stores/thinkdb/      # Think database layer
app/domain/thinkapp/thinkapp.go               # Think handlers
app/domain/thinkapp/route.go                  # Think routes
api/services/partners/main.go                 # Service initialization
api/services/partners/all/all.go              # Route registration
app/sdk/mux/mux.go                            # Mux configuration
app/sdk/mid/authen.go                         # Auth middleware
zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem  # RSA key
```

### Frontend
```
frontend/lib/auth-context.tsx                 # Auth state management
frontend/app/login/page.tsx                   # Login page
frontend/components/auth/ProtectedRoute.tsx   # Route protection
frontend/app/layout.tsx                       # Root layout
frontend/components/layout/Header.tsx         # Header with auth UI
frontend/app/thinks/page.tsx                  # Protected thinks page
frontend/lib/api.ts                           # API client with auth
frontend/lib/types.ts                         # TypeScript types
```

---

## API Endpoints

### Authentication
```
GET /v1/auth/token/{kid}
Headers: Authorization: Basic base64(email:password)
Response: {"token":"eyJhbG..."}
```

### Thinks (Authenticated)
```
GET /v1/thinks
Headers: Authorization: Bearer <token>
Response: {"items":[...],"total":10,"page":1,"rowsPerPage":10}

POST /v1/thinks
Headers: Authorization: Bearer <token>
Body: {"category":"personal","content":"text"}
Response: {"id":"...","category":"personal",...}

GET /v1/thinks/{id}
Headers: Authorization: Bearer <token>
Response: {"id":"...","category":"personal",...}
```

---

## Environment Variables

### Backend (Local)
```bash
# .env
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=db
```

### Backend (Hetzner)
```bash
# .env
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=<strong-password>
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.lat,https://*.vercel.app
```

### Frontend (Vercel)
```bash
NEXT_PUBLIC_API_URL=https://api.rafiki.lat
```

---

## Troubleshooting

### Backend won't start
```bash
docker compose logs partner-service | grep -i error
# Check: "key not found" → verify RSA key exists
# Check: migration errors → verify SQL syntax
```

### Login fails (401)
```bash
# Verify user exists
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT email, enabled FROM users;"

# Verify password hash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT email, password_hash FROM users WHERE email='test@example.com';"
```

### Token not sent from frontend
```javascript
// Browser console
localStorage.getItem('rafiki_auth_token')
// Should not be null

// Check Network tab for Authorization header
```

### CORS errors
```bash
# Check backend CORS configuration
cat /opt/rafiki/.env | grep CORS

# Should include frontend origin:
# https://app.rafiki.lat or https://*.vercel.app
```

---

## Success Checklist

### Backend
- [ ] All 11 bugs fixed
- [ ] RSA key generated and loaded
- [ ] Service starts without errors
- [ ] Auth endpoints accessible
- [ ] Test user can login
- [ ] Token expiry is 48 hours
- [ ] Thinks require authentication
- [ ] Users only see their own thinks

### Frontend
- [ ] Login page displays
- [ ] Form validation works
- [ ] Successful login redirects
- [ ] Token stored in localStorage
- [ ] Protected routes work
- [ ] Header shows user info
- [ ] Logout clears auth state
- [ ] API requests include Bearer token

### Integration
- [ ] End-to-end login works
- [ ] Think creation associates with user
- [ ] Think queries are user-scoped
- [ ] No CORS errors
- [ ] Token persists across refreshes
- [ ] 401 triggers auto-logout

---

## Timeline

**Day 1 Morning (4-6 hours):**
- Fix all backend bugs
- Generate RSA keys
- Test locally

**Day 1 Afternoon (2-3 hours):**
- Deploy to Hetzner
- Create production user
- Verify endpoints

**Day 2 Morning (4-6 hours):**
- Create frontend files
- Modify existing files
- Test locally

**Day 2 Afternoon (2-3 hours):**
- Deploy to Vercel
- End-to-end testing
- Bug fixes

**Total: 12-18 hours (1-2 days)**

---

## Key Decisions Summary

1. **User Creation:** SQL only (no registration endpoint for now)
2. **Token Expiry:** 48 hours
3. **Token Storage:** localStorage (frontend)
4. **RSA Keys:** Dev key in repo, prod key on server
5. **Database:** Wipe and start fresh (confirmed by user)
6. **Frontend Scope:** Login only (no registration page)

---

## Next Features (Future)

- User registration endpoint (POST /v1/users)
- User profile page
- Password change functionality
- Forgot password flow
- Admin user management UI
- Token refresh mechanism
- Two-factor authentication

---

**Ready to Start?**
1. Begin with Backend Plan: [BACKEND_AUTH_IMPLEMENTATION_PLAN.md](./BACKEND_AUTH_IMPLEMENTATION_PLAN.md)
2. Then Frontend Plan: [FRONTEND_AUTH_IMPLEMENTATION_PLAN.md](./FRONTEND_AUTH_IMPLEMENTATION_PLAN.md)
3. Refer to this document for quick lookups
