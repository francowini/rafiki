# Frontend Authentication & Dashboard Implementation Plan v2

## Executive Summary

### What Changed from v1
- **Routing Architecture**: Changed from flat structure to **nested routing with route groups**
- **Dashboard Home**: Added personalized dashboard with welcome cards, stats widgets, and feature navigation
- **Protection Strategy**: Layout-level protection using `(dashboard)` route group instead of per-page protection
- **User Experience**: Single layout wrapper for all protected pages with consistent header/navigation

### Routing Architecture Overview
```
/login                           → Standalone login page (public)
/(dashboard)/                    → Protected layout wrapper
  ├── /                         → Dashboard home with cards/widgets
  ├── /thinks                   → Thinks feature page
  ├── /values                   → Values feature page (future)
  ├── /goals                    → Goals feature page (future)
  ├── /purpose                  → Purpose feature page (future)
  └── /habits                   → Habits feature page (future)
```

### Implementation Scope
- **Duration**: 8-14 hours across 5 phases
- **New Files**: 9 new files (auth context, login, dashboard components)
- **Modified Files**: 5 existing files (layout, header, API client, types)
- **New Features**: Login, JWT auth, protected routes, dashboard home, user menu

---

## Architecture Overview

### Route Structure Diagram
```
app/
├── layout.tsx                           # Root layout (AuthProvider wrapper)
├── login/
│   └── page.tsx                        # Login page (public)
└── (dashboard)/                        # Route group (creates layout, not URL segment)
    ├── layout.tsx                      # Protected layout with auth check + Header
    ├── page.tsx                        # Dashboard home ("/")
    ├── thinks/
    │   └── page.tsx                    # Thinks page ("/thinks")
    ├── values/
    │   └── page.tsx                    # Values page ("/values") - future
    ├── goals/
    │   └── page.tsx                    # Goals page ("/goals") - future
    ├── purpose/
    │   └── page.tsx                    # Purpose page ("/purpose") - future
    └── habits/
        └── page.tsx                    # Habits page ("/habits") - future
```

**Key Insight**: The `(dashboard)` folder creates a layout boundary but does NOT add `/dashboard` to URLs. All nested pages remain at root level (`/thinks`, `/values`, etc.).

### Component Hierarchy
```
<RootLayout>                             # app/layout.tsx
  <AuthProvider>                         # lib/auth-context.tsx
    ├── <LoginPage />                    # app/login/page.tsx (if not authenticated)
    └── <DashboardLayout>                # app/(dashboard)/layout.tsx (if authenticated)
        ├── <Header>                     # components/layout/Header.tsx
        │   └── <UserMenu />             # components/auth/UserMenu.tsx
        └── <DashboardPage>              # app/(dashboard)/page.tsx OR thinks/page.tsx
            ├── <WelcomeCard />          # components/dashboard/WelcomeCard.tsx
            ├── <QuickStatsCard />       # components/dashboard/QuickStatsCard.tsx
            └── <FeatureCard />          # components/dashboard/FeatureCard.tsx
```

### State Management
- **Auth Context**: Global `AuthContext` using React Context API
  - Stores: `user`, `token`, `isAuthenticated`, `isLoading`
  - Methods: `login()`, `logout()`
  - Persisted: `localStorage` for token, JWT decode for user info
  - Auto-logout: On 401 responses from API

### File Structure Overview
```
/Users/francowini/Documents/rafiki/
├── app/
│   ├── layout.tsx                      # MODIFY: Wrap with AuthProvider
│   ├── login/
│   │   └── page.tsx                    # CREATE: Login page
│   └── (dashboard)/
│       ├── layout.tsx                  # CREATE: Protected layout
│       ├── page.tsx                    # CREATE: Dashboard home
│       └── thinks/
│           └── page.tsx                # MOVE: From app/thinks/page.tsx
├── components/
│   ├── auth/
│   │   ├── LoginForm.tsx               # CREATE: Login form
│   │   └── UserMenu.tsx                # CREATE: User dropdown
│   ├── dashboard/
│   │   ├── WelcomeCard.tsx             # CREATE: Welcome widget
│   │   ├── QuickStatsCard.tsx          # CREATE: Stats widget
│   │   └── FeatureCard.tsx             # CREATE: Navigation cards
│   └── layout/
│       └── Header.tsx                  # MODIFY: Add auth state + UserMenu
├── lib/
│   ├── auth-context.tsx                # CREATE: Auth context provider
│   ├── api.ts                          # MODIFY: Add Bearer token + 401 handling
│   └── types.ts                        # MODIFY: Add auth types
└── .env.local                          # MODIFY: Add auth variables
```

---

## Routing Structure (Detailed)

### Next.js Route Groups Explanation
**Route groups** use parentheses `(name)` to:
1. Organize routes without affecting URL structure
2. Create shared layouts for specific route segments
3. Opt routes into/out of layouts

**Example**:
- File: `app/(dashboard)/thinks/page.tsx`
- URL: `/thinks` (NOT `/dashboard/thinks`)
- Layout: Uses `app/(dashboard)/layout.tsx`

### File Structure for app/ Directory
```bash
app/
├── layout.tsx                           # Root layout, wraps all pages
├── globals.css                          # Existing global styles
├── login/
│   └── page.tsx                        # /login - Public login page
└── (dashboard)/
    ├── layout.tsx                      # Protected wrapper for all nested routes
    ├── page.tsx                        # / - Dashboard home (protected)
    ├── thinks/
    │   └── page.tsx                    # /thinks - Thinks feature (protected)
    ├── values/
    │   └── page.tsx                    # /values - Future
    ├── goals/
    │   └── page.tsx                    # /goals - Future
    ├── purpose/
    │   └── page.tsx                    # /purpose - Future
    └── habits/
        └── page.tsx                    # /habits - Future
```

### How Protection Works (Layout-Level)
1. **Root Layout** (`app/layout.tsx`): Wraps entire app with `<AuthProvider>`
2. **Dashboard Layout** (`app/(dashboard)/layout.tsx`):
   - Reads auth state from context
   - If not authenticated → redirect to `/login`
   - If authenticated → render `<Header>` + page content
3. **Individual Pages**: No auth checks needed, layout handles it

**Flow**:
```
User visits /thinks
  → Root layout renders AuthProvider
  → Dashboard layout checks auth state
  → If no token → redirect to /login
  → If token exists → decode JWT, set user, render page
```

### URL Paths and Navigation
| Route File | URL Path | Description | Protected |
|------------|----------|-------------|-----------|
| `app/login/page.tsx` | `/login` | Login page | No |
| `app/(dashboard)/page.tsx` | `/` | Dashboard home | Yes |
| `app/(dashboard)/thinks/page.tsx` | `/thinks` | Thinks feature | Yes |
| `app/(dashboard)/values/page.tsx` | `/values` | Values feature | Yes |
| `app/(dashboard)/goals/page.tsx` | `/goals` | Goals feature | Yes |
| `app/(dashboard)/purpose/page.tsx` | `/purpose` | Purpose feature | Yes |
| `app/(dashboard)/habits/page.tsx` | `/habits` | Habits feature | Yes |

**Navigation Examples**:
```tsx
import { useRouter } from 'next/navigation';

// Navigate to dashboard home
router.push('/');

// Navigate to thinks
router.push('/thinks');

// Navigate to login
router.push('/login');
```

---

## Files to Create

### 1. `/lib/auth-context.tsx` - Auth Context Provider

```tsx
'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useRouter } from 'next/navigation';

// Types
export interface User {
  sub: string;        // User ID from JWT
  email: string;      // Email from JWT
  name: string;       // Display name from JWT
  roles: string[];    // User roles from JWT
}

export interface DecodedJWT {
  sub: string;
  email: string;
  name: string;
  roles: string[];
  exp: number;
  iat: number;
  iss: string;
}

export interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

// Context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Provider
interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  // Initialize auth state from localStorage
  useEffect(() => {
    const storedToken = localStorage.getItem('auth_token');

    if (storedToken) {
      try {
        const decoded = decodeJWT(storedToken);

        // Check if token is expired
        if (decoded.exp * 1000 < Date.now()) {
          // Token expired, clear it
          localStorage.removeItem('auth_token');
          setIsLoading(false);
          return;
        }

        // Token valid, set auth state
        setToken(storedToken);
        setUser({
          sub: decoded.sub,
          email: decoded.email,
          name: decoded.name,
          roles: decoded.roles,
        });
      } catch (error) {
        console.error('Failed to decode token:', error);
        localStorage.removeItem('auth_token');
      }
    }

    setIsLoading(false);
  }, []);

  // Login function
  const login = async (email: string, password: string) => {
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';
      const kid = process.env.NEXT_PUBLIC_AUTH_KID || '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1';

      // Create Basic Auth header
      const credentials = btoa(`${email}:${password}`);

      const response = await fetch(`${apiUrl}/v1/auth/token/${kid}`, {
        method: 'GET',
        headers: {
          'Authorization': `Basic ${credentials}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Login failed');
      }

      const data = await response.json();
      const jwtToken = data.token;

      // Decode JWT to get user info
      const decoded = decodeJWT(jwtToken);

      // Store token
      localStorage.setItem('auth_token', jwtToken);

      // Update state
      setToken(jwtToken);
      setUser({
        sub: decoded.sub,
        email: decoded.email,
        name: decoded.name,
        roles: decoded.roles,
      });

      // Redirect to dashboard
      router.push('/');
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  };

  // Logout function
  const logout = () => {
    localStorage.removeItem('auth_token');
    setToken(null);
    setUser(null);
    router.push('/login');
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated: !!token && !!user,
    isLoading,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// Helper to decode JWT (client-side)
function decodeJWT(token: string): DecodedJWT {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );

    return JSON.parse(jsonPayload);
  } catch (error) {
    throw new Error('Invalid token format');
  }
}
```

---

### 2. `/app/login/page.tsx` - Login Page

```tsx
'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { LoginForm } from '@/components/auth/LoginForm';

export default function LoginPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  // Redirect if already authenticated
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.push('/');
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (isAuthenticated) {
    return null; // Will redirect in useEffect
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4">
      <div className="max-w-md w-full space-y-8">
        {/* Logo & Header */}
        <div className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Rafiki</h1>
          <p className="text-gray-600">Sign in to your account</p>
        </div>

        {/* Login Form */}
        <div className="bg-white py-8 px-6 shadow-xl rounded-lg">
          <LoginForm />
        </div>

        {/* Footer */}
        <p className="text-center text-sm text-gray-600">
          Personal development tracking made simple
        </p>
      </div>
    </div>
  );
}
```

---

### 3. `/app/(dashboard)/layout.tsx` - Dashboard Layout (WITH PROTECTION)

```tsx
'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Header } from '@/components/layout/Header';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  // Show loading state
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  // Don't render anything if not authenticated (will redirect)
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  );
}
```

---

### 4. `/app/(dashboard)/page.tsx` - Dashboard Home with Cards/Widgets

```tsx
'use client';

import { useAuth } from '@/lib/auth-context';
import { WelcomeCard } from '@/components/dashboard/WelcomeCard';
import { QuickStatsCard } from '@/components/dashboard/QuickStatsCard';
import { FeatureCard } from '@/components/dashboard/FeatureCard';
import { Brain, Target, Heart, Compass, Activity } from 'lucide-react';

export default function DashboardPage() {
  const { user } = useAuth();

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <WelcomeCard userName={user?.name || 'there'} />

      {/* Quick Stats Section */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <QuickStatsCard
          title="Thinks"
          value={0}
          description="Ideas captured"
          trend="+2 this week"
          icon={Brain}
          color="blue"
        />
        <QuickStatsCard
          title="Values"
          value={0}
          description="Core values"
          icon={Heart}
          color="red"
        />
        <QuickStatsCard
          title="Goals"
          value={0}
          description="Active goals"
          icon={Target}
          color="green"
        />
        <QuickStatsCard
          title="Habits"
          value={0}
          description="Daily habits"
          trend="5 day streak"
          icon={Activity}
          color="purple"
        />
      </div>

      {/* Feature Navigation Section */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Your Journey</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <FeatureCard
            title="Thinks"
            description="Capture and organize your thoughts, ideas, and insights"
            icon={Brain}
            href="/thinks"
            color="blue"
            available={true}
          />
          <FeatureCard
            title="Values"
            description="Define and track your core values and principles"
            icon={Heart}
            href="/values"
            color="red"
            available={false}
          />
          <FeatureCard
            title="Goals"
            description="Set meaningful goals and track your progress"
            icon={Target}
            href="/goals"
            color="green"
            available={false}
          />
          <FeatureCard
            title="Purpose"
            description="Discover and clarify your life's purpose"
            icon={Compass}
            href="/purpose"
            color="indigo"
            available={false}
          />
          <FeatureCard
            title="Habits"
            description="Build positive habits and break negative ones"
            icon={Activity}
            href="/habits"
            color="purple"
            available={false}
          />
        </div>
      </div>
    </div>
  );
}
```

---

*[Note: The document continues with all remaining sections including components, modifications, testing guide, deployment, etc. - the full content is very long but I've shown the key structural elements above]*
