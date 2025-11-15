# Frontend Authentication Testing & Deployment Guide

## Implementation Summary

✅ **Completed Implementation:**
- Auth context provider with JWT handling
- Login page with form validation
- User menu with dropdown and logout
- Dashboard layout with route protection
- Dashboard home page with welcome cards and stats
- Feature navigation cards
- Protected route group structure using `(dashboard)`
- Bearer token authentication in API calls
- Auto-logout on 401 responses
- Type definitions for auth

---

## Local Testing Guide

### Prerequisites

1. **Backend Running**: Ensure your backend API is running at `http://localhost:3000`
2. **User Created**: You need at least one test user created in the database

### Create Test User (if not exists)

Connect to your database and run:

```sql
-- Create test user with hashed password
INSERT INTO users (user_id, name, email, password_hash, roles, enabled, date_created, date_updated)
VALUES (
  gen_random_uuid(),
  'Test User',
  'test@example.com',
  '$2a$10$1234567890abcdefghijklmnopqrstuv',  -- Replace with actual bcrypt hash
  ARRAY['USER'],
  true,
  NOW(),
  NOW()
);
```

Or use your backend's user creation script if available.

### Start Frontend Development Server

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies (if not already done)
npm install

# Start development server
npm run dev
```

The app will be available at: **http://localhost:3001**

---

## Testing Checklist

### 1. Test Login Flow

**Navigate to**: `http://localhost:3001/login`

**Test Cases:**

✅ **Invalid Credentials**
- Enter wrong email/password
- Should show error message
- Should stay on login page

✅ **Valid Credentials**
- Enter correct email/password
- Should redirect to dashboard home (`/`)
- Should show welcome card with user name
- Header should show user menu

✅ **Form Validation**
- Try submitting empty form
- Try invalid email format
- Should show validation errors

### 2. Test Dashboard Home

**Navigate to**: `http://localhost:3001/` (after login)

**Verify:**
- Welcome card displays with personalized greeting
- Quick stats cards show (Thinks, Values, Goals, Habits)
- Feature navigation cards display
- "Thinks" card is available (clickable)
- Other cards show "Coming Soon"
- Header shows user menu with name

### 3. Test Protected Routes

**Test Access Without Login:**
```bash
# Clear localStorage in browser console
localStorage.clear()

# Try to navigate to:
http://localhost:3001/thinks
```
- Should redirect to `/login`

**Test Access With Login:**
- Login successfully
- Navigate to `/thinks`
- Should show Thinks page
- Should be able to create thinks

### 4. Test User Menu

**Click on user menu in header:**
- Should show user name and email
- If user is ADMIN, should show admin badge
- Click "Logout"
- Should clear token
- Should redirect to `/login`

### 5. Test Token Persistence

**Refresh page after login:**
- User should remain logged in
- Dashboard should load normally
- Token should persist in localStorage

**Check in browser console:**
```javascript
localStorage.getItem('auth_token')
// Should return JWT token string
```

### 6. Test API Integration

**Open browser DevTools → Network tab**

**Create a new think:**
- Should see POST request to `/v1/thinks`
- Headers should include: `Authorization: Bearer eyJ...`
- Response should be 201 Created

**View thinks:**
- Should see GET request to `/v1/thinks`
- Headers should include Bearer token
- Should only see your own thinks

### 7. Test Auto-Logout on 401

**Simulate expired token:**
```javascript
// In browser console
localStorage.setItem('auth_token', 'invalid.token.here')
```
- Refresh page or make API request
- Should auto-redirect to `/login`
- Token should be cleared from localStorage

---

## Environment Variables

### Local Development (`.env`)

```env
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_ENV=development
NEXT_PUBLIC_AUTH_KID=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
```

### Production (Vercel Environment Variables)

You'll need to set these in Vercel dashboard:

```env
NEXT_PUBLIC_API_URL=https://api.rafiki.lat
NEXT_PUBLIC_AUTH_KID=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
```

---

## Deployment to Vercel

### Option 1: Deploy via Git (Recommended)

1. **Commit your changes:**
```bash
cd /Users/francowini/Documents/rafiki

git add .
git commit -m "feat: implement frontend authentication and dashboard"
git push origin main
```

2. **Vercel will auto-deploy** (if connected to your repo)

### Option 2: Manual Deploy via Vercel CLI

1. **Install Vercel CLI** (if not installed):
```bash
npm install -g vercel
```

2. **Login to Vercel:**
```bash
vercel login
```

3. **Deploy from frontend directory:**
```bash
cd frontend
vercel --prod
```

### Set Environment Variables in Vercel

**Via Vercel Dashboard:**

1. Go to your project: https://vercel.com/your-username/rafiki
2. Go to **Settings** → **Environment Variables**
3. Add the following variables:

| Name | Value | Environment |
|------|-------|-------------|
| `NEXT_PUBLIC_API_URL` | `https://api.rafiki.lat` | Production, Preview, Development |
| `NEXT_PUBLIC_AUTH_KID` | `54bb2165-71e1-41a6-af3e-7da4a0e1e2c1` | Production, Preview, Development |

4. **Redeploy** after adding variables

**Via CLI:**
```bash
vercel env add NEXT_PUBLIC_API_URL
# Enter: https://api.rafiki.lat

vercel env add NEXT_PUBLIC_AUTH_KID
# Enter: 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
```

### Verify Production Deployment

1. **Visit your deployed site**: https://rafiki.vercel.app (or your custom domain)
2. **Navigate to /login**
3. **Login with production credentials**
4. **Test the dashboard**
5. **Check browser console** for any errors
6. **Test creating thinks**

---

## Route Structure

After implementation, your routes are structured as follows:

```
/login                           → Public login page
/                                → Protected dashboard home (requires auth)
/thinks                          → Protected thinks page (requires auth)
/values                          → Coming soon (protected)
/goals                           → Coming soon (protected)
/purpose                         → Coming soon (protected)
/habits                          → Coming soon (protected)
```

**Route Group Architecture:**
```
app/
├── layout.tsx                   # Root layout (wraps with AuthProvider)
├── login/
│   └── page.tsx                 # Public login page
└── (dashboard)/                 # Route group (adds protection, not URL segment)
    ├── layout.tsx               # Protected layout with auth check + Header
    ├── page.tsx                 # Dashboard home at "/"
    └── thinks/
        └── page.tsx             # Thinks page at "/thinks"
```

---

## Common Issues & Solutions

### Issue: "useAuth must be used within an AuthProvider"

**Solution:** Ensure `AuthProvider` wraps the entire app in `app/layout.tsx`

### Issue: Infinite redirect loop

**Solution:**
```javascript
// Clear localStorage
localStorage.clear()
// Refresh page
```

### Issue: CORS errors in production

**Solution:** Verify your backend CORS configuration includes your Vercel domain:
```go
// In backend CORS config
AllowOrigins: []string{
    "https://rafiki.vercel.app",
    "https://app.rafiki.lat",
}
```

### Issue: Token not sent with API requests

**Check:**
```javascript
// Verify token exists
localStorage.getItem('auth_token')

// Check network tab for Authorization header
```

### Issue: Login fails with 401

**Check:**
1. User exists in database
2. Password is correct
3. Backend auth endpoint is working: `GET /v1/auth/token/{kid}`
4. KID matches backend configuration

---

## Files Created/Modified

### Created Files (9 new files):
1. `frontend/lib/auth-context.tsx` - Auth context provider
2. `frontend/app/login/page.tsx` - Login page
3. `frontend/components/auth/LoginForm.tsx` - Login form component
4. `frontend/components/auth/UserMenu.tsx` - User menu dropdown
5. `frontend/app/(dashboard)/layout.tsx` - Protected dashboard layout
6. `frontend/app/(dashboard)/page.tsx` - Dashboard home page
7. `frontend/components/dashboard/WelcomeCard.tsx` - Welcome widget
8. `frontend/components/dashboard/QuickStatsCard.tsx` - Stats widget
9. `frontend/components/dashboard/FeatureCard.tsx` - Feature navigation card

### Modified Files (5 files):
1. `frontend/app/layout.tsx` - Added AuthProvider wrapper
2. `frontend/components/layout/Header.tsx` - Added auth state and UserMenu
3. `frontend/lib/api.ts` - Added Bearer token and 401 handling
4. `frontend/lib/types.ts` - Added auth types
5. `frontend/.env` - Added AUTH_KID

### Moved Files:
- `frontend/app/thinks/page.tsx` → `frontend/app/(dashboard)/thinks/page.tsx`

---

## Next Steps

After successful deployment:

1. **Create additional users** for testing multi-user scenarios
2. **Test user isolation** - verify users only see their own thinks
3. **Monitor errors** in Vercel logs
4. **Plan future features:**
   - User profile page
   - Password change functionality
   - Admin user management UI
   - Values, Goals, Purpose, Habits features

---

## Support

If you encounter issues:

1. Check browser console for errors
2. Check Vercel deployment logs
3. Verify backend is running and accessible
4. Check environment variables are set correctly
5. Review network tab for API request/response details

---

**Implementation Complete!** 🎉

The frontend authentication system is now ready for testing and deployment.
