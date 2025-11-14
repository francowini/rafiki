# Frontend Authentication Implementation Plan

**Status:** Ready for Implementation (After Backend Deploy)
**Estimated Time:** 4-6 hours
**Scope:** Login only (no registration page)
**Token Storage:** localStorage
**Token Expiry:** 48 hours

---

## Overview

This plan covers frontend changes to implement login functionality. Users will:
1. Navigate to `/login` page
2. Enter email and password
3. Receive JWT token from backend
4. Store token in localStorage
5. Access protected routes with Bearer token
6. See only their own thinks

**User Creation:** Admin creates users via SQL (no self-registration)

---

## Architecture

### Authentication Flow

```
┌─────────────┐
│ Login Page  │
│ (email/pw)  │
└──────┬──────┘
       │
       ├─ POST Basic Auth
       │  GET /v1/auth/token/{kid}
       ▼
┌─────────────┐
│   Backend   │
│ Returns JWT │
└──────┬──────┘
       │
       ├─ Store in localStorage
       ▼
┌─────────────┐
│  Protected  │
│   Routes    │
│ (Thinks)    │
└─────────────┘
```

### State Management

- **Auth Context** (React Context API)
  - Stores: user data, token, isAuthenticated
  - Methods: login(), logout()
  - Persists across page refreshes

- **localStorage Keys**
  - `rafiki_auth_token` - JWT token string
  - `rafiki_user` - User data JSON (decoded from token)

---

## Files to Create

### 1. Auth Context Provider
**Path:** `frontend/lib/auth-context.tsx`

```typescript
"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface User {
  user_id: string;
  email: string;
  roles: string[];
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// KID for JWT signing (matches backend key)
const DEFAULT_KID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1";
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";

// Decode JWT payload (without verification - for client-side display only)
function decodeJWT(token: string): any {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;

    const payload = parts[1];
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
    return decoded;
  } catch {
    return null;
  }
}

// Check if token is expired
function isTokenExpired(token: string): boolean {
  const decoded = decodeJWT(token);
  if (!decoded || !decoded.exp) return true;

  const now = Math.floor(Date.now() / 1000);
  return decoded.exp < now;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Initialize auth state from localStorage on mount
  useEffect(() => {
    const initAuth = () => {
      const storedToken = localStorage.getItem('rafiki_auth_token');
      const storedUser = localStorage.getItem('rafiki_user');

      if (storedToken && !isTokenExpired(storedToken)) {
        setToken(storedToken);

        if (storedUser) {
          try {
            setUser(JSON.parse(storedUser));
          } catch (error) {
            console.error('Failed to parse stored user', error);
            logout();
          }
        }
      } else if (storedToken) {
        // Token exists but is expired
        console.log('Token expired, clearing auth state');
        logout();
      }

      setIsLoading(false);
    };

    initAuth();
  }, []);

  const login = async (email: string, password: string) => {
    try {
      // Create Basic Auth header
      const credentials = btoa(`${email}:${password}`);

      // Call backend auth endpoint
      const response = await fetch(
        `${API_BASE_URL}/v1/auth/token/${DEFAULT_KID}`,
        {
          method: 'GET',
          headers: {
            'Authorization': `Basic ${credentials}`,
          },
        }
      );

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Login failed: ${response.status} ${errorText}`);
      }

      const data = await response.json();
      const newToken = data.token;

      if (!newToken) {
        throw new Error('No token received from server');
      }

      // Decode token to extract user info
      const decoded = decodeJWT(newToken);
      if (!decoded) {
        throw new Error('Invalid token format');
      }

      // Create user object from token claims
      const userData: User = {
        user_id: decoded.sub,
        email: email, // Store email since it's not in token
        roles: decoded.roles || [],
      };

      // Store in localStorage
      localStorage.setItem('rafiki_auth_token', newToken);
      localStorage.setItem('rafiki_user', JSON.stringify(userData));

      // Update state
      setToken(newToken);
      setUser(userData);
    } catch (error) {
      // Clear any partial state
      logout();
      throw error;
    }
  };

  const logout = () => {
    localStorage.removeItem('rafiki_auth_token');
    localStorage.removeItem('rafiki_user');
    setToken(null);
    setUser(null);
  };

  const value: AuthContextType = {
    user,
    token,
    isLoading,
    isAuthenticated: !!token && !!user,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
```

---

### 2. Login Page
**Path:** `frontend/app/login/page.tsx`

```typescript
"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(1, "Password is required"),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const { login, isAuthenticated, isLoading: authLoading } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  // Redirect if already authenticated
  useEffect(() => {
    if (!authLoading && isAuthenticated) {
      router.push("/thinks");
    }
  }, [isAuthenticated, authLoading, router]);

  // Show nothing while checking auth
  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  const onSubmit = async (data: LoginFormData) => {
    setError(null);
    setIsSubmitting(true);

    try {
      await login(data.email, data.password);
      // Navigation happens via useEffect above
    } catch (err: any) {
      console.error('Login error:', err);
      setError(
        err.message || "Login failed. Please check your credentials and try again."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold">Login to Rafiki</CardTitle>
          <CardDescription>
            Enter your email and password to access your thinks
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="your@email.com"
                autoComplete="email"
                {...register("email")}
              />
              {errors.email && (
                <p className="text-sm text-red-500">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                autoComplete="current-password"
                {...register("password")}
              />
              {errors.password && (
                <p className="text-sm text-red-500">{errors.password.message}</p>
              )}
            </div>

            <Button type="submit" disabled={isSubmitting} className="w-full">
              {isSubmitting ? (
                <span className="flex items-center justify-center">
                  <span className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></span>
                  Logging in...
                </span>
              ) : (
                "Login"
              )}
            </Button>
          </form>

          <div className="mt-4 text-center text-sm text-gray-600">
            <p>Don't have an account? Contact your administrator.</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
```

---

### 3. Protected Route Component
**Path:** `frontend/components/auth/ProtectedRoute.tsx`

```typescript
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, isLoading, router]);

  // Show loading state while checking auth
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  // Don't render children if not authenticated (redirect happening)
  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
```

---

## Files to Modify

### 4. Root Layout - Wrap with AuthProvider
**Path:** `frontend/app/layout.tsx`

**Add import:**
```typescript
import { AuthProvider } from "@/lib/auth-context";
```

**Wrap children:**
```typescript
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <AuthProvider>
          <div className="min-h-screen bg-background">
            <Header />
            <main className="container mx-auto py-6 px-4">
              {children}
            </main>
          </div>
        </AuthProvider>
      </body>
    </html>
  );
}
```

---

### 5. Update Header Component
**Path:** `frontend/components/layout/Header.tsx`

```typescript
"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth-context";

export function Header() {
  const { isAuthenticated, user, logout } = useAuth();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <header className="border-b">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" className="text-2xl font-bold">
            Rafiki
          </Link>
          {isAuthenticated && (
            <nav className="flex gap-4">
              <Link href="/thinks">
                <Button variant="ghost">Thinks</Button>
              </Link>
            </nav>
          )}
        </div>
        <div className="flex items-center gap-4">
          {isAuthenticated && user ? (
            <>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">
                  {user.email}
                </span>
                {user.roles.includes('ADMIN') && (
                  <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                    Admin
                  </span>
                )}
              </div>
              <Button variant="outline" onClick={handleLogout}>
                Logout
              </Button>
            </>
          ) : (
            <Link href="/login">
              <Button>Login</Button>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
```

---

### 6. Update Thinks Page - Add Protection
**Path:** `frontend/app/thinks/page.tsx`

**Add import:**
```typescript
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
```

**Wrap page content:**
```typescript
export default function ThinksPage() {
  const [refreshKey, setRefreshKey] = useState(0);

  const handleThinkCreated = () => {
    setRefreshKey((prev) => prev + 1);
  };

  return (
    <ProtectedRoute>
      <div className="space-y-8">
        <div className="max-w-2xl">
          <ThinkForm onSuccess={handleThinkCreated} />
        </div>

        <div>
          <ThinkList refresh={refreshKey} />
        </div>
      </div>
    </ProtectedRoute>
  );
}
```

---

### 7. Update API Client - Add Auth Header
**Path:** `frontend/lib/api.ts`

**Update fetchAPI function:**
```typescript
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";

async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  // Get token from localStorage
  const token = typeof window !== 'undefined'
    ? localStorage.getItem('rafiki_auth_token')
    : null;

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options?.headers,
  };

  // Add Bearer token if available
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    });

    // Handle 401 Unauthorized - clear auth and redirect
    if (response.status === 401) {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('rafiki_auth_token');
        localStorage.removeItem('rafiki_user');
        window.location.href = '/login';
      }
      throw new APIError(401, 'Unauthorized - please login again');
    }

    // Handle 403 Forbidden
    if (response.status === 403) {
      throw new APIError(403, 'Forbidden - insufficient permissions');
    }

    if (!response.ok) {
      const errorText = await response.text();
      throw new APIError(
        response.status,
        `API request failed: ${response.status} ${errorText}`
      );
    }

    return response.json();
  } catch (error) {
    if (error instanceof APIError) {
      throw error;
    }
    throw new Error(
      `Network error: ${error instanceof Error ? error.message : "Unknown error"}`
    );
  }
}
```

---

### 8. Update Types
**Path:** `frontend/lib/types.ts`

**Add auth types:**
```typescript
// User types
export interface User {
  user_id: string;
  email: string;
  roles: string[];
}

export interface AuthTokenResponse {
  token: string;
}

export interface DecodedToken {
  sub: string;        // user_id
  iss: string;        // issuer
  exp: number;        // expiration timestamp
  iat: number;        // issued at timestamp
  roles: string[];    // user roles
}
```

---

## Environment Variables

### Development
**Path:** `frontend/.env.local`

```env
NEXT_PUBLIC_API_URL=http://localhost:3000
```

### Production
**Path:** Vercel Environment Variables

```env
NEXT_PUBLIC_API_URL=https://api.rafiki.lat
```

---

## Manual Testing Guide

### Prerequisites

1. Backend deployed and running
2. Test user created via SQL
3. Backend accessible at configured API_URL

### Step 1: Start Frontend

```bash
cd frontend

# Install dependencies (if not done)
npm install

# Start dev server
npm run dev

# Open browser
open http://localhost:3000
```

### Step 2: Test Login Page Access

**Actions:**
1. Navigate to http://localhost:3000/login
2. Verify login form displays
3. Verify "Email" and "Password" fields present
4. Verify "Login" button present

**Expected:**
- Clean login page with Rafiki branding
- No errors in console
- Form fields are empty

---

### Step 3: Test Form Validation

**Test 3a: Empty Form**
1. Click "Login" button without entering anything

**Expected:**
- Email error: "Invalid email address"
- Password error: "Password is required"
- No API request sent

**Test 3b: Invalid Email**
1. Enter "notanemail" in email field
2. Enter "password123" in password field
3. Click "Login"

**Expected:**
- Email error: "Invalid email address"
- No API request sent

**Test 3c: Valid Format**
1. Enter "test@example.com" in email
2. Enter "password123" in password
3. Click "Login"

**Expected:**
- Form errors clear
- API request sent
- Button shows "Logging in..." with spinner

---

### Step 4: Test Invalid Login

**Test 4a: Wrong Password**
1. Enter valid user email: "test@example.com"
2. Enter wrong password: "wrongpassword"
3. Click "Login"

**Expected:**
- Red alert appears: "Login failed: 401 Unauthorized"
- Form remains visible
- No token stored in localStorage
- User stays on login page

**Verify in DevTools:**
```javascript
// Open console
localStorage.getItem('rafiki_auth_token')
// Expected: null
```

**Test 4b: Non-existent User**
1. Enter email: "fake@example.com"
2. Enter any password
3. Click "Login"

**Expected:**
- Error message displays
- No token stored
- User stays on login page

---

### Step 5: Test Successful Login

**Actions:**
1. Enter valid credentials (from SQL user creation)
   - Email: `test@example.com`
   - Password: `password123` (or whatever you set)
2. Click "Login"

**Expected:**
1. Button shows "Logging in..." with spinner
2. After ~1 second, redirect to `/thinks`
3. Header shows user email
4. Header shows "Logout" button

**Verify in DevTools:**
```javascript
// Open browser console
localStorage.getItem('rafiki_auth_token')
// Expected: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

localStorage.getItem('rafiki_user')
// Expected: {"user_id":"...","email":"test@example.com","roles":["USER"]}
```

**Verify Network Request:**
1. Open DevTools Network tab
2. Look for request to `/v1/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1`
3. Check request headers:
   - `Authorization: Basic dGVzdEBleGFtcGxlLmNvbTpwYXNzd29yZDEyMw==`
4. Check response:
   - Status: `200 OK`
   - Body: `{"token":"eyJhbG..."}`

---

### Step 6: Test Protected Routes

**Test 6a: Access Thinks While Logged In**
1. Navigate to http://localhost:3000/thinks
2. Verify page loads without redirect

**Expected:**
- Thinks page displays
- Think form visible
- Can create thinks
- Header shows user info

**Test 6b: Access Thinks Without Login**
1. Clear localStorage:
   ```javascript
   localStorage.clear()
   ```
2. Navigate to http://localhost:3000/thinks

**Expected:**
- Immediately redirected to `/login`
- Login form displays
- No thinks content visible

---

### Step 7: Test Token Persistence

**Actions:**
1. Login successfully
2. Refresh the page (F5 or Cmd+R)

**Expected:**
- Page reloads
- User remains logged in
- Token still in localStorage
- Header still shows user info
- No redirect to login

---

### Step 8: Test Logout

**Actions:**
1. While logged in, click "Logout" button in header

**Expected:**
1. Token removed from localStorage
2. User redirected to `/login`
3. Header shows "Login" button instead of user info

**Verify in DevTools:**
```javascript
localStorage.getItem('rafiki_auth_token')
// Expected: null

localStorage.getItem('rafiki_user')
// Expected: null
```

---

### Step 9: Test Auto-Redirect When Authenticated

**Actions:**
1. Login successfully (now at `/thinks`)
2. Navigate to http://localhost:3000/login in URL bar

**Expected:**
- Immediately redirected back to `/thinks`
- Cannot access login page while authenticated

---

### Step 10: Test API Requests Include Token

**Actions:**
1. Login successfully
2. Open DevTools Network tab
3. Create a new think

**Expected Network Request:**
```
POST /v1/thinks
Headers:
  Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
  Content-Type: application/json
Body:
  {"category":"personal","content":"test think"}
```

**Response:**
```
Status: 201 Created
Body: {"id":"...","category":"personal","content":"test think",...}
```

---

### Step 11: Test 401 Auto-Logout

**Actions:**
1. Login successfully
2. Open DevTools Console
3. Manually expire the token:
   ```javascript
   // Set an expired token
   localStorage.setItem('rafiki_auth_token', 'expired.token.here')
   ```
4. Try to query thinks (refresh page or make API call)

**Expected:**
- API returns 401
- Token cleared from localStorage
- Redirected to `/login`
- Alert/error message shows

---

### Step 12: Test Multiple Users

**Setup:**
1. Create second user via SQL (see backend plan)
2. Logout from first user

**Test:**
1. Login as user2@example.com
2. Create a think
3. Logout
4. Login as test@example.com
5. View thinks list

**Expected:**
- User2 only sees their own think
- User1 only sees their own thinks
- No overlap between users' thinks

---

### Step 13: Test Token Expiry Time

**Actions:**
1. Login successfully
2. Decode token in console:
   ```javascript
   const token = localStorage.getItem('rafiki_auth_token');
   const parts = token.split('.');
   const payload = JSON.parse(atob(parts[1]));
   console.log('Issued at:', new Date(payload.iat * 1000));
   console.log('Expires at:', new Date(payload.exp * 1000));
   console.log('Duration (hours):', (payload.exp - payload.iat) / 3600);
   ```

**Expected Output:**
```
Issued at: 2025-11-12T10:00:00.000Z
Expires at: 2025-11-14T10:00:00.000Z
Duration (hours): 48
```

---

### Step 14: Test CORS

**Actions:**
1. Open browser console on deployed frontend (Vercel)
2. Test API call:
   ```javascript
   fetch('https://api.rafiki.lat/v1/readiness')
     .then(r => r.json())
     .then(console.log)
     .catch(console.error)
   ```

**Expected:**
- Request succeeds
- No CORS errors in console
- Response: `{"status":"ok"}`

**Test with Auth:**
```javascript
const token = localStorage.getItem('rafiki_auth_token');
fetch('https://api.rafiki.lat/v1/thinks', {
  headers: { 'Authorization': `Bearer ${token}` }
})
  .then(r => r.json())
  .then(console.log)
  .catch(console.error)
```

**Expected:**
- Request succeeds
- Authorization header sent
- Response: thinks array

---

### Step 15: Test Error Handling

**Test 15a: Network Error**
1. Stop backend service
2. Try to login

**Expected:**
- Error message: "Network error: ..."
- User stays on login page
- No token stored

**Test 15b: Invalid Response**
1. Backend returns 500 error
2. Try to create think

**Expected:**
- Error message displayed
- User remains logged in (not logged out for 5xx errors)

---

## Testing Checklist

### Login Page
- [ ] Login page displays at `/login`
- [ ] Email field validates format
- [ ] Password field is required
- [ ] Form validation works client-side
- [ ] Invalid credentials show error
- [ ] Valid credentials redirect to `/thinks`
- [ ] Loading state shows during login
- [ ] Error messages are clear and helpful

### Authentication State
- [ ] Token stored in localStorage after login
- [ ] User data stored in localStorage
- [ ] Token persists across page refreshes
- [ ] Auto-logout on expired token
- [ ] Logout clears localStorage
- [ ] Logout redirects to login

### Protected Routes
- [ ] `/thinks` requires authentication
- [ ] Unauthenticated users redirected to login
- [ ] Loading state shown during auth check
- [ ] Authenticated users can access thinks
- [ ] Cannot access login page when authenticated

### API Integration
- [ ] Authorization header sent on all requests
- [ ] Bearer token format correct
- [ ] 401 responses trigger auto-logout
- [ ] Thinks are user-scoped (only see own)
- [ ] Create think associates with logged-in user

### User Experience
- [ ] Header shows user email when logged in
- [ ] Header shows logout button when logged in
- [ ] Header shows login button when logged out
- [ ] Admin badge shows for admin users
- [ ] No flash of unauthenticated content
- [ ] Smooth transitions between auth states

### Error Handling
- [ ] Network errors show user-friendly message
- [ ] Invalid credentials show clear error
- [ ] 401 errors trigger logout and redirect
- [ ] 403 errors show permission denied
- [ ] Form validation errors display correctly

---

## Common Issues and Solutions

### Issue: Infinite redirect loop

**Symptoms:** Browser keeps redirecting between `/login` and `/thinks`

**Solutions:**
```javascript
// Check if token exists but is invalid
const token = localStorage.getItem('rafiki_auth_token');
console.log('Token:', token);

// Clear localStorage and try again
localStorage.clear();
```

### Issue: CORS errors

**Symptoms:**
```
Access to fetch at 'http://localhost:3000/v1/auth/token/...' from origin 'http://localhost:3001' has been blocked by CORS
```

**Solutions:**
1. Check `NEXT_PUBLIC_API_URL` is correct
2. Verify backend CORS configuration includes frontend origin
3. Check for preflight OPTIONS request in Network tab

### Issue: 401 even with valid token

**Symptoms:** API returns 401 despite valid-looking token

**Solutions:**
```javascript
// Decode and inspect token
const token = localStorage.getItem('rafiki_auth_token');
const parts = token.split('.');
const payload = JSON.parse(atob(parts[1]));
console.log('Token payload:', payload);

// Check expiry
const now = Math.floor(Date.now() / 1000);
console.log('Token expired:', payload.exp < now);

// Verify token format
console.log('Token parts:', parts.length); // Should be 3
```

### Issue: Token not sent with requests

**Symptoms:** Network tab shows requests without Authorization header

**Solutions:**
1. Verify token exists in localStorage
2. Check `fetchAPI` function includes token injection
3. Verify requests use the `fetchAPI` wrapper

```javascript
// Test token retrieval
const token = localStorage.getItem('rafiki_auth_token');
console.log('Token exists:', !!token);
```

### Issue: User stays logged in after token expires

**Symptoms:** User can access protected routes after 48 hours

**Solutions:**
- Token expiry check happens on API call, not page load
- Make an API request to trigger expiry check
- Add expiry check to auth context initialization

---

## Deployment to Vercel

### Step 1: Set Environment Variables

In Vercel Dashboard:
1. Go to Project Settings → Environment Variables
2. Add:
   ```
   NEXT_PUBLIC_API_URL=https://api.rafiki.lat
   ```
3. Set for: Production, Preview, Development

### Step 2: Deploy

```bash
cd frontend

# Option 1: Auto-deploy (Git integration)
git add .
git commit -m "feat: Implement login authentication"
git push origin main

# Option 2: Manual deploy
vercel --prod
```

### Step 3: Test Production

1. Visit deployed URL (e.g., https://rafiki.vercel.app)
2. Navigate to `/login`
3. Login with production credentials
4. Verify thinks are user-scoped
5. Test logout
6. Check browser console for errors

---

## Success Criteria

✅ Login page displays and validates input
✅ Users can login with credentials
✅ JWT token stored in localStorage
✅ Token sent as Bearer in API requests
✅ Protected routes redirect to login
✅ Logout clears auth state
✅ Token persists across refreshes
✅ 401 errors trigger auto-logout
✅ Users only see their own thinks
✅ No CORS errors
✅ Clean error messages for users

---

## Next Steps

After frontend implementation:
1. Deploy to Vercel
2. Test end-to-end with production backend
3. Create additional test users
4. Plan future features:
   - User profile page
   - Password change
   - User registration (when backend ready)
   - Admin user management UI

---

**Questions?** Review the multi-mind analysis document for detailed architecture explanation.
