# Frontend Implementation Plan - Rafiki Thinks Management

## Executive Summary

This document outlines the complete implementation plan for the Rafiki "thinks" management frontend, a Next.js application deployed on Vercel that integrates with the existing Go backend on Hetzner.

### Current State
- ✅ Go backend with all required API endpoints fully implemented
- ✅ PostgreSQL database with proper schema
- ✅ Docker Compose deployment on Hetzner CPX11
- ⚠️ CORS configured but not wired to mux (1 critical fix needed)
- ❌ No Nginx reverse proxy (direct port 3000 exposure)
- ❌ No SSL/TLS on backend
- ❌ No frontend application

### Target State
- ✅ Next.js 14+ frontend on Vercel
- ✅ Nginx reverse proxy with SSL/TLS
- ✅ CORS properly configured for production
- ✅ Domain strategy: `api.rafiki.com` + `app.rafiki.com`
- ✅ Complete CI/CD pipeline
- ✅ Monitoring and alerting

### Estimated Timeline
- **Phase 1 (Backend Prep):** 1 hour
- **Phase 2 (Infrastructure):** 2-3 hours
- **Phase 3 (Frontend Dev):** 6-8 hours
- **Phase 4 (Integration):** 2 hours
- **Phase 5 (Production):** 1 hour
- **Total:** 12-15 hours (1-2 days)

### Critical Dependencies
1. CORS fix must be deployed before frontend development
2. Nginx setup must be complete before DNS configuration
3. SSL certificates must be obtained before production deployment

---

## Architecture Overview

### Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         INTERNET                                 │
└────────────────────────┬────────────────────────────────────────┘
                         │
         ┌───────────────┴───────────────┐
         │                               │
         ▼                               ▼
┌────────────────────┐          ┌────────────────────┐
│   VERCEL CDN       │          │  HETZNER SERVER    │
│  (Edge Network)    │          │  178.156.170.37    │
└────────────────────┘          └────────────────────┘
         │                               │
         │                               │
┌────────▼────────────┐         ┌────────▼────────────┐
│  Next.js Frontend   │         │   Nginx Reverse     │
│                     │         │      Proxy          │
│  app.rafiki.com     │──────────▶  (Port 80/443)     │
│                     │  HTTPS   │                     │
│  - shadcn/ui        │  Calls   │  - SSL/TLS          │
│  - Tailwind CSS     │          │  - Rate Limiting    │
│  - TypeScript       │          │  - CORS             │
└─────────────────────┘          └─────────┬───────────┘
                                           │
                                 ┌─────────▼───────────┐
                                 │  Docker Network     │
                                 │  10.10.0.0/24       │
                                 └─────────┬───────────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    │                      │                      │
           ┌────────▼────────┐    ┌───────▼────────┐   ┌────────▼─────────┐
           │ partner-service │    │   PostgreSQL   │   │  Tempo/Grafana   │
           │   (Port 3000)   │    │   (Port 5432)  │   │  (Observability) │
           │                 │    │                │   │                  │
           │  - Go API       │    │  - Thinks DB   │   │  - Tracing       │
           │  - CORS enabled │    │  - Migrations  │   │  - Metrics       │
           └─────────────────┘    └────────────────┘   └──────────────────┘
```

### Domain Strategy

- **Frontend:** `app.rafiki.com` (Vercel)
- **Backend API:** `api.rafiki.com` (Hetzner + Nginx)
- **Observability:** `grafana.rafiki.com` (Hetzner + Nginx, optional)

### Technology Stack

**Frontend:**
- Framework: Next.js 14+ (App Router)
- UI: shadcn/ui components
- Styling: Tailwind CSS
- Language: TypeScript
- Deployment: Vercel
- State Management: React hooks (no external state lib for MVP)

**Backend (Existing):**
- Language: Go 1.22+
- Framework: Standard library (http.ServeMux)
- Database: PostgreSQL 16
- Logging: Structured logging (slog)
- Observability: Tempo + Grafana

**Infrastructure:**
- Reverse Proxy: Nginx
- SSL: Let's Encrypt (via Certbot)
- Containerization: Docker Compose
- Server: Hetzner CPX11 (2 vCPU, 2GB RAM)

---

## Backend Readiness Assessment

### ✅ API Endpoints (Fully Implemented)

#### GET /v1/thinks
**Location:** [api/services/partner/mux/mux.go](../api/services/partner/mux/mux.go)
**Handler:** [app/domain/thinkapp/thinkapp.go:48-72](../app/domain/thinkapp/thinkapp.go)

**Features:**
- Pagination via query params: `page`, `rows`, `orderBy`
- Default pagination: page=1, rows=10
- Supported order fields: `think_id`, `category`, `date_created`, `date_updated`

**Response Format:**
```json
{
  "items": [
    {
      "id": "uuid-string",
      "category": "personal",
      "content": "Think content here",
      "dateCreated": "2024-01-01T00:00:00Z",
      "dateUpdated": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "rowsPerPage": 10
}
```

#### GET /v1/thinks/{think_id}
**Location:** [app/domain/thinkapp/thinkapp.go:75-87](../app/domain/thinkapp/thinkapp.go)

**Features:**
- UUID validation
- 404 handling for not found
- Single think object response

#### POST /v1/thinks
**Location:** [app/domain/thinkapp/thinkapp.go:27-45](../app/domain/thinkapp/thinkapp.go)

**Request Body:**
```json
{
  "category": "personal",
  "content": "Think content here"
}
```

**Features:**
- Category validation (enum: personal, work, ideas, learning, reflection)
- Content validation (non-empty)
- Auto-generated UUID and timestamps
- Returns created think object

### ✅ Data Models (Exact Match)

**Categories:** [business/domain/thinkbus/model.go:18-26](../business/domain/thinkbus/model.go)
```go
const (
    CategoryPersonal   Category = "personal"     // ✅
    CategoryWork       Category = "work"         // ✅
    CategoryIdeas      Category = "ideas"        // ✅
    CategoryLearning   Category = "learning"     // ✅
    CategoryReflection Category = "reflection"   // ✅
)
```

**Think Model:** [app/domain/thinkapp/model.go:14-21](../app/domain/thinkapp/model.go)
```go
type Think struct {
    ID          string `json:"id"`
    Category    string `json:"category"`
    Content     string `json:"content"`
    DateCreated string `json:"dateCreated"`  // RFC3339 format
    DateUpdated string `json:"dateUpdated"`  // RFC3339 format
}
```

### ⚠️ CORS Configuration (Needs Fix)

**Status:** CORS is configured but not wired to the mux.

**Current Configuration:** [api/services/partners/main.go:76](../api/services/partners/main.go)
```go
CORSAllowedOrigins []string `conf:"default:*"`
```

**Docker Compose:** [docker-compose.yml:144](../docker-compose.yml)
```yaml
- PARTNER_WEB_CORSALLOWEDORIGINS=*
```

**CORS Implementation:** [foundation/web/web.go:68-91](../foundation/web/web.go)
- Sets `Access-Control-Allow-Origin` header
- Sets `Access-Control-Allow-Methods`: POST, PATCH, GET, OPTIONS, PUT, DELETE
- Sets `Access-Control-Allow-Headers`: Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization

**Problem:** The `WithCORS` option is not passed to `mux.WebAPI`

**Location to Fix:** [api/services/partners/main.go:220-222](../api/services/partners/main.go)

---

## Implementation Timeline

### Week 1: Preparation & Infrastructure

| Day | Phase | Tasks | Duration |
|-----|-------|-------|----------|
| 1 | Backend Prep | Fix CORS, test API endpoints | 1 hour |
| 1-2 | Infrastructure | Set up Nginx, obtain SSL, configure DNS | 2-3 hours |
| 2-3 | Frontend Dev | Create Next.js project, setup shadcn/ui | 2 hours |
| 3-4 | Frontend Dev | Implement components (ThinkForm, ThinkList) | 4-6 hours |
| 4 | Integration | Test frontend with backend API | 2 hours |
| 5 | Production Deploy | Deploy to Vercel, final testing | 1 hour |

### Week 2: Monitoring & Optimization

| Day | Phase | Tasks |
|-----|-------|-------|
| 1-2 | Monitoring | Set up uptime monitoring, configure alerts |
| 3 | Testing | Load testing, CORS testing, error handling |
| 4-5 | Documentation | API docs, runbook, deployment guide |

---


## Phase 3: Frontend Development

**Duration:** 6-8 hours
**Prerequisites:** Backend CORS fix deployed
**Blocker for:** Integration testing

### Task 3.1: Create Next.js Project

**Steps:**

```bash
# Navigate to frontend directory
cd /Users/francowini/Documents/rafiki
mkdir -p frontend
cd frontend

# Create Next.js project
npx create-next-app@latest . \
  --typescript \
  --tailwind \
  --app \
  --no-src-dir \
  --import-alias "@/*"

# Accept defaults:
# ✔ Would you like to use ESLint? › Yes
# ✔ Would you like to use Turbopack for development? › No
# ✔ Would you like to customize the import alias (@/* by default)? › No

# Initialize git
git init
git add .
git commit -m "chore: initialize Next.js project"
```

### Task 3.2: Setup shadcn/ui

**Steps:**

```bash
# Initialize shadcn/ui
npx shadcn@latest init

# Accept defaults:
# ✔ Which style would you like to use? › New York
# ✔ Which color would you like to use as base color? › Zinc
# ✔ Do you want to use CSS variables for colors? › Yes

# Install required components
npx shadcn@latest add button
npx shadcn@latest add card
npx shadcn@latest add input
npx shadcn@latest add textarea
npx shadcn@latest add select
npx shadcn@latest add form
npx shadcn@latest add dialog
npx shadcn@latest add badge
npx shadcn@latest add skeleton
npx shadcn@latest add alert

# Commit changes
git add .
git commit -m "chore: setup shadcn/ui and install components"
```

### Task 3.3: Create Folder Structure

**Create the following folders:**

```bash
# Create folder structure
mkdir -p lib
mkdir -p components/features
mkdir -p components/layout
mkdir -p app/thinks

# Verify structure
tree -L 2 .
```

**Expected structure:**

```
frontend/
├── app/
│   ├── layout.tsx
│   ├── page.tsx
│   └── thinks/
│       └── page.tsx
├── components/
│   ├── ui/              # shadcn/ui components
│   ├── features/
│   │   ├── ThinkForm.tsx
│   │   ├── ThinkCard.tsx
│   │   └── ThinkList.tsx
│   └── layout/
│       ├── Header.tsx
│       └── Sidebar.tsx
├── lib/
│   ├── api.ts
│   ├── types.ts
│   └── utils.ts (from shadcn)
├── public/
└── config files (tsconfig.json, tailwind.config.ts, etc.)
```

### Task 3.4: Create TypeScript Types

**File:** `lib/types.ts`

```typescript
// lib/types.ts

export type ThinkCategory = "personal" | "work" | "ideas" | "learning" | "reflection";

export interface Think {
  id: string;
  category: ThinkCategory;
  content: string;
  dateCreated: string;
  dateUpdated: string;
}

export interface NewThink {
  category: ThinkCategory;
  content: string;
}

export interface ThinkListResponse {
  items: Think[];
  total: number;
  page: number;
  rowsPerPage: number;
}

export interface PaginationParams {
  page?: number;
  rows?: number;
  orderBy?: "think_id" | "category" | "date_created" | "date_updated";
}
```

### Task 3.5: Create API Client

**File:** `lib/api.ts`

```typescript
// lib/api.ts
import { Think, ThinkListResponse, NewThink, PaginationParams } from "./types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";

class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "APIError";
  }
}

async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

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
    throw new Error(`Network error: ${error instanceof Error ? error.message : "Unknown error"}`);
  }
}

export const api = {
  thinks: {
    /**
     * Get all thinks with pagination
     */
    getAll: async (params?: PaginationParams): Promise<ThinkListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.set("page", params.page.toString());
      if (params?.rows) queryParams.set("rows", params.rows.toString());
      if (params?.orderBy) queryParams.set("orderBy", params.orderBy);

      const query = queryParams.toString();
      const endpoint = `/v1/thinks${query ? `?${query}` : ""}`;

      return fetchAPI<ThinkListResponse>(endpoint);
    },

    /**
     * Get a single think by ID
     */
    getById: async (id: string): Promise<Think> => {
      return fetchAPI<Think>(`/v1/thinks/${id}`);
    },

    /**
     * Create a new think
     */
    create: async (data: NewThink): Promise<Think> => {
      return fetchAPI<Think>("/v1/thinks", {
        method: "POST",
        body: JSON.stringify(data),
      });
    },
  },

  health: {
    /**
     * Check API readiness
     */
    readiness: async (): Promise<{ status: string }> => {
      return fetchAPI<{ status: string }>("/v1/readiness");
    },

    /**
     * Check API liveness
     */
    liveness: async (): Promise<{ status: string }> => {
      return fetchAPI<{ status: string }>("/v1/liveness");
    },
  },
};

export { APIError };
```

### Task 3.6: Create Main Layout

**File:** `app/layout.tsx`

```typescript
// app/layout.tsx
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Header } from "@/components/layout/Header";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Rafiki Thinks - Thought Management",
  description: "Manage your thoughts, ideas, and reflections with Rafiki",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <div className="min-h-screen bg-background">
          <Header />
          <main className="container mx-auto py-6 px-4">
            {children}
          </main>
        </div>
      </body>
    </html>
  );
}
```

### Task 3.7: Create Header Component

**File:** `components/layout/Header.tsx`

```typescript
// components/layout/Header.tsx
import Link from "next/link";
import { Button } from "@/components/ui/button";

export function Header() {
  return (
    <header className="border-b">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" className="text-2xl font-bold">
            Rafiki Thinks
          </Link>
          <nav className="flex gap-4">
            <Link href="/thinks">
              <Button variant="ghost">Thinks</Button>
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">
            {process.env.NEXT_PUBLIC_ENV || "development"}
          </span>
        </div>
      </div>
    </header>
  );
}
```

### Task 3.8: Create ThinkCard Component

**File:** `components/features/ThinkCard.tsx`

```typescript
// components/features/ThinkCard.tsx
import { Think } from "@/lib/types";
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface ThinkCardProps {
  think: Think;
}

const categoryColors: Record<Think["category"], string> = {
  personal: "bg-blue-500",
  work: "bg-green-500",
  ideas: "bg-purple-500",
  learning: "bg-yellow-500",
  reflection: "bg-pink-500",
};

export function ThinkCard({ think }: ThinkCardProps) {
  const formattedDate = new Date(think.dateCreated).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <Badge className={categoryColors[think.category]}>
            {think.category}
          </Badge>
          <span className="text-xs text-muted-foreground">{formattedDate}</span>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm whitespace-pre-wrap">{think.content}</p>
      </CardContent>
      <CardFooter className="pt-2">
        <span className="text-xs text-muted-foreground">ID: {think.id.slice(0, 8)}</span>
      </CardFooter>
    </Card>
  );
}
```

### Task 3.9: Create ThinkForm Component

**File:** `components/features/ThinkForm.tsx`

```typescript
// components/features/ThinkForm.tsx
"use client";

import { useState } from "react";
import { NewThink, ThinkCategory } from "@/lib/types";
import { api, APIError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";

const categories: ThinkCategory[] = [
  "personal",
  "work",
  "ideas",
  "learning",
  "reflection",
];

interface ThinkFormProps {
  onSuccess?: () => void;
}

export function ThinkForm({ onSuccess }: ThinkFormProps) {
  const [category, setCategory] = useState<ThinkCategory>("personal");
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!content.trim()) {
      setError("Content cannot be empty");
      return;
    }

    setIsSubmitting(true);

    try {
      const newThink: NewThink = { category, content };
      await api.thinks.create(newThink);

      // Reset form
      setContent("");
      setCategory("personal");

      // Call success callback
      if (onSuccess) {
        onSuccess();
      }
    } catch (err) {
      if (err instanceof APIError) {
        setError(`Failed to create think: ${err.message}`);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create New Think</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <label htmlFor="category" className="text-sm font-medium">
              Category
            </label>
            <Select value={category} onValueChange={(v) => setCategory(v as ThinkCategory)}>
              <SelectTrigger>
                <SelectValue placeholder="Select a category" />
              </SelectTrigger>
              <SelectContent>
                {categories.map((cat) => (
                  <SelectItem key={cat} value={cat}>
                    {cat.charAt(0).toUpperCase() + cat.slice(1)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <label htmlFor="content" className="text-sm font-medium">
              Content
            </label>
            <Textarea
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="What's on your mind?"
              rows={5}
              className="resize-none"
            />
          </div>

          <Button type="submit" disabled={isSubmitting} className="w-full">
            {isSubmitting ? "Creating..." : "Create Think"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
```

### Task 3.10: Create ThinkList Component

**File:** `components/features/ThinkList.tsx`

```typescript
// components/features/ThinkList.tsx
"use client";

import { useEffect, useState } from "react";
import { ThinkListResponse } from "@/lib/types";
import { api, APIError } from "@/lib/api";
import { ThinkCard } from "./ThinkCard";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface ThinkListProps {
  refresh?: number; // Increment to trigger refresh
}

export function ThinkList({ refresh = 0 }: ThinkListProps) {
  const [data, setData] = useState<ThinkListResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);

  const fetchThinks = async (pageNum: number) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.thinks.getAll({
        page: pageNum,
        rows: 10,
        orderBy: "date_created",
      });
      setData(response);
    } catch (err) {
      if (err instanceof APIError) {
        setError(`Failed to load thinks: ${err.message}`);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchThinks(page);
  }, [page, refresh]);

  if (isLoading && !data) {
    return (
      <div className="space-y-4">
        {[...Array(3)].map((_, i) => (
          <Skeleton key={i} className="h-32 w-full" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (!data || data.items.length === 0) {
    return (
      <Alert>
        <AlertDescription>
          No thinks yet. Create your first think above!
        </AlertDescription>
      </Alert>
    );
  }

  const totalPages = Math.ceil(data.total / data.rowsPerPage);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">
          All Thinks ({data.total})
        </h2>
        <span className="text-sm text-muted-foreground">
          Page {data.page} of {totalPages}
        </span>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data.items.map((think) => (
          <ThinkCard key={think.id} think={think} />
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1 || isLoading}
          >
            Previous
          </Button>
          <span className="text-sm">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages || isLoading}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
```

### Task 3.11: Create Thinks Page

**File:** `app/thinks/page.tsx`

```typescript
// app/thinks/page.tsx
"use client";

import { useState } from "react";
import { ThinkForm } from "@/components/features/ThinkForm";
import { ThinkList } from "@/components/features/ThinkList";

export default function ThinksPage() {
  const [refreshKey, setRefreshKey] = useState(0);

  const handleThinkCreated = () => {
    // Trigger refresh of ThinkList
    setRefreshKey((prev) => prev + 1);
  };

  return (
    <div className="space-y-8">
      <div className="max-w-2xl">
        <ThinkForm onSuccess={handleThinkCreated} />
      </div>

      <div>
        <ThinkList refresh={refreshKey} />
      </div>
    </div>
  );
}
```

### Task 3.12: Create Home Page

**File:** `app/page.tsx`

```typescript
// app/page.tsx
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function HomePage() {
  return (
    <div className="space-y-8">
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-bold">Welcome to Rafiki Thinks</h1>
        <p className="text-xl text-muted-foreground">
          Manage your thoughts, ideas, and reflections
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 max-w-4xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle>Personal</CardTitle>
            <CardDescription>Personal thoughts and reflections</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Work</CardTitle>
            <CardDescription>Work-related notes and ideas</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Ideas</CardTitle>
            <CardDescription>Creative ideas and inspiration</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Learning</CardTitle>
            <CardDescription>Learning notes and insights</CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Reflection</CardTitle>
            <CardDescription>Deep reflections and analysis</CardDescription>
          </CardHeader>
        </Card>
      </div>

      <div className="text-center">
        <Link href="/thinks">
          <Button size="lg">Get Started</Button>
        </Link>
      </div>
    </div>
  );
}
```

### Task 3.13: Create Environment Variables

**File:** `.env.local`

```bash
# .env.local (for local development)
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_ENV=development
```

**File:** `.env.example`

```bash
# .env.example
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_ENV=development
```

### Task 3.14: Test Locally

**Steps:**

```bash
# Start the development server
npm run dev

# Open browser to http://localhost:3000

# Test the following:
# 1. Home page loads
# 2. Navigate to /thinks
# 3. Create a new think
# 4. Verify it appears in the list
# 5. Test pagination (if you have more than 10 thinks)
# 6. Test different categories
```

### Acceptance Criteria

- [ ] Next.js project created with TypeScript and Tailwind
- [ ] shadcn/ui installed and configured
- [ ] All components implemented (ThinkForm, ThinkCard, ThinkList)
- [ ] API client working with error handling
- [ ] Home page and Thinks page functional
- [ ] Can create new thinks
- [ ] Can view list of thinks
- [ ] Pagination works
- [ ] Responsive design (mobile, tablet, desktop)
- [ ] Loading states implemented
- [ ] Error handling implemented

---

## Phase 4: Deployment & Integration

**Duration:** 2 hours
**Prerequisites:** Frontend development complete, Infrastructure setup complete
**Blocker for:** Production deployment

### Task 4.1: Set Up Vercel Project

**Steps:**

```bash
# Install Vercel CLI
npm install -g vercel

# Login to Vercel
vercel login

# Navigate to frontend directory
cd /Users/francowini/Documents/rafiki/frontend

# Initialize Vercel project
vercel init

# Link to existing project or create new one
vercel link

# Follow prompts:
# ? Set up and deploy "~/Documents/rafiki/frontend"? Yes
# ? Which scope do you want to deploy to? (select your account)
# ? Link to existing project? No
# ? What's your project's name? rafiki-frontend
# ? In which directory is your code located? ./
```

### Task 4.2: Configure Vercel Environment Variables

**Via Vercel Dashboard:**

```bash
# Go to: https://vercel.com/dashboard
# Select project: rafiki-frontend
# Go to: Settings → Environment Variables

# Add the following variables:

# Production Environment
Variable: NEXT_PUBLIC_API_URL
Value: https://api.rafiki.com
Environment: Production

Variable: NEXT_PUBLIC_ENV
Value: production
Environment: Production

# Preview Environment
Variable: NEXT_PUBLIC_API_URL
Value: https://api.rafiki.com  (or staging URL if you have one)
Environment: Preview

Variable: NEXT_PUBLIC_ENV
Value: preview
Environment: Preview

# Development Environment (optional)
Variable: NEXT_PUBLIC_API_URL
Value: http://localhost:3000
Environment: Development

Variable: NEXT_PUBLIC_ENV
Value: development
Environment: Development
```

**Via Vercel CLI:**

```bash
# Set production environment variable
vercel env add NEXT_PUBLIC_API_URL production
# Enter: https://api.rafiki.com

vercel env add NEXT_PUBLIC_ENV production
# Enter: production

# Set preview environment variable
vercel env add NEXT_PUBLIC_API_URL preview
# Enter: https://api.rafiki.com

vercel env add NEXT_PUBLIC_ENV preview
# Enter: preview

# Verify environment variables
vercel env ls
```

### Task 4.3: Deploy to Vercel Preview

**Steps:**

```bash
# From frontend directory
cd /Users/francowini/Documents/rafiki/frontend

# Deploy to preview
vercel

# This will:
# 1. Build your Next.js app
# 2. Deploy to a preview URL (e.g., rafiki-frontend-xxx.vercel.app)
# 3. Return the preview URL

# Test the preview deployment
# Open the URL in your browser
# Verify:
# - Home page loads
# - Navigate to /thinks
# - Try creating a new think
# - Verify CORS works
# - Check API connectivity
```

### Task 4.4: Test CORS Integration

**Steps:**

```bash
# 1. Open browser DevTools (F12)
# 2. Go to Network tab
# 3. Navigate to /thinks page on preview URL
# 4. Create a new think

# 5. Verify the following in Network tab:
# - OPTIONS request to https://api.rafiki.com/v1/thinks
#   - Should return 200 OK
#   - Should have Access-Control-Allow-Origin header
# - POST request to https://api.rafiki.com/v1/thinks
#   - Should return 201 Created
#   - Should have Access-Control-Allow-Origin header

# 6. If CORS fails:
# - Check backend CORS configuration in .env
# - Verify Vercel preview URL is in CORS allowed origins
# - Check Nginx configuration
```

### Task 4.5: Update Backend CORS for Vercel Preview

**If CORS fails, update backend configuration:**

```bash
# SSH into Hetzner server
ssh root@178.156.170.37

# Edit .env file
nano /opt/rafiki/.env

# Update CORS to include Vercel preview domains
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.com,https://rafiki-frontend-*.vercel.app,http://localhost:3000

# Note: The wildcard (*) in vercel.app should work for all preview deployments

# Restart services
docker compose down
docker compose up -d

# Verify CORS
curl -i -X OPTIONS https://api.rafiki.com/v1/thinks \
  -H "Origin: https://rafiki-frontend-xxx.vercel.app" \
  -H "Access-Control-Request-Method: POST"

# Expected: Access-Control-Allow-Origin header in response
```

### Task 4.6: Performance Testing

**Steps:**

```bash
# 1. Test page load speed
# Use Lighthouse in Chrome DevTools
# Target: Performance score > 90

# 2. Test API response times
# Open Network tab
# Create multiple thinks
# Verify response times < 500ms

# 3. Test mobile responsiveness
# Use Chrome DevTools device emulation
# Test on: iPhone SE, iPhone 12, iPad, Desktop

# 4. Test pagination
# Create 15+ thinks
# Verify pagination works correctly
# Verify page transitions are smooth

# 5. Test error handling
# Disconnect internet
# Try creating a think
# Verify error message is user-friendly
```

### Acceptance Criteria

- [ ] Vercel project created and linked
- [ ] Environment variables configured
- [ ] Preview deployment successful
- [ ] CORS working between Vercel and Hetzner
- [ ] Can create thinks from preview URL
- [ ] Can view thinks from preview URL
- [ ] Pagination works
- [ ] Mobile responsive
- [ ] Performance score > 90
- [ ] Error handling works

---

## Phase 5: Production Deployment

**Duration:** 1 hour
**Prerequisites:** Integration testing complete, all tests passing
**Blocker for:** None (final phase)

### Task 5.1: Configure Custom Domain in Vercel

**Steps:**

```bash
# 1. Go to Vercel Dashboard
# https://vercel.com/dashboard/rafiki-frontend

# 2. Go to Settings → Domains
# 3. Click "Add Domain"
# 4. Enter: app.rafiki.com
# 5. Vercel will provide DNS instructions:
#    - CNAME record: app.rafiki.com → cname.vercel-dns.com

# 6. Go to your DNS provider (e.g., Namecheap, Cloudflare)
# 7. Add the CNAME record as instructed
# 8. Wait for DNS propagation (5-30 minutes)

# 9. Verify DNS propagation
dig app.rafiki.com
# Should return CNAME record

# 10. In Vercel dashboard, click "Refresh" to verify domain
# 11. Vercel will automatically provision SSL certificate
```

### Task 5.2: Update Backend CORS for Production Domain

**Update backend to allow production domain:**

```bash
# SSH into Hetzner server
ssh root@178.156.170.37

# Edit .env file
nano /opt/rafiki/.env

# Update CORS to production domain
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.com,https://rafiki-frontend-*.vercel.app,http://localhost:3000

# Restart services
docker compose down
docker compose up -d

# Verify CORS
curl -i -X OPTIONS https://api.rafiki.com/v1/thinks \
  -H "Origin: https://app.rafiki.com" \
  -H "Access-Control-Request-Method: POST"

# Expected: Access-Control-Allow-Origin: https://app.rafiki.com
```

### Task 5.3: Deploy to Production

**Steps:**

```bash
# From frontend directory
cd /Users/francowini/Documents/rafiki/frontend

# Commit all changes
git add .
git commit -m "feat: complete frontend implementation"
git push origin main

# Deploy to production
vercel --prod

# This will:
# 1. Build your Next.js app
# 2. Deploy to production (app.rafiki.com)
# 3. Return the production URL

# Verify deployment
echo "Production URL: https://app.rafiki.com"
```

### Task 5.4: Production Testing

**Complete testing checklist:**

```bash
# 1. Open https://app.rafiki.com
# 2. Verify home page loads
# 3. Navigate to /thinks
# 4. Create a new think
# 5. Verify it appears in the list
# 6. Test pagination (create 15+ thinks if needed)
# 7. Test all 5 categories
# 8. Test on mobile devices
# 9. Test on different browsers (Chrome, Firefox, Safari, Edge)
# 10. Verify SSL certificate is valid (green padlock)

# 11. Check Network tab:
# - All API calls to https://api.rafiki.com
# - CORS headers present
# - No 404 or 500 errors
# - Response times < 500ms

# 12. Check Console tab:
# - No JavaScript errors
# - No CORS errors
# - No network errors
```

### Task 5.5: Set Up Monitoring

**UptimeRobot (Free):**

```bash
# 1. Go to https://uptimerobot.com
# 2. Create free account
# 3. Add monitors:

# Monitor 1: Frontend
# Type: HTTPS
# URL: https://app.rafiki.com
# Interval: 5 minutes
# Alert contacts: your-email@example.com

# Monitor 2: Backend API
# Type: HTTPS
# URL: https://api.rafiki.com/v1/readiness
# Interval: 5 minutes
# Alert contacts: your-email@example.com

# Monitor 3: Backend Liveness
# Type: HTTPS
# URL: https://api.rafiki.com/v1/liveness
# Interval: 5 minutes
# Alert contacts: your-email@example.com
```

**Vercel Analytics (Optional):**

```bash
# 1. Go to Vercel Dashboard → Analytics
# 2. Enable Analytics for rafiki-frontend project
# 3. Add @vercel/analytics package to frontend:

cd /Users/francowini/Documents/rafiki/frontend
npm install @vercel/analytics

# 4. Update app/layout.tsx:
# import { Analytics } from '@vercel/analytics/react';
# Add <Analytics /> component to layout

# 5. Deploy update:
vercel --prod
```

### Task 5.6: Update Documentation

**Create DEPLOYMENT.md in frontend folder:**

```bash
cd /Users/francowini/Documents/rafiki/frontend
nano DEPLOYMENT.md
```

**Content:**

```markdown
# Deployment Guide - Rafiki Frontend

## Production URLs

- Frontend: https://app.rafiki.com
- Backend API: https://api.rafiki.com
- Grafana: https://grafana.rafiki.com (optional)

## Environment Variables

### Vercel (Frontend)
- `NEXT_PUBLIC_API_URL`: https://api.rafiki.com
- `NEXT_PUBLIC_ENV`: production

### Hetzner (Backend)
- `PARTNER_WEB_CORSALLOWEDORIGINS`: https://app.rafiki.com,https://rafiki-frontend-*.vercel.app

## Deployment Process

### Frontend (Automatic)
1. Push to main branch
2. Vercel auto-deploys to production
3. Monitor deployment: https://vercel.com/dashboard

### Backend (Manual)
1. SSH into server: `ssh root@178.156.170.37`
2. Run deployment script: `/opt/rafiki/devops/deploy.sh`
3. Verify health: `curl https://api.rafiki.com/v1/readiness`

## Rollback

### Frontend
1. Go to Vercel Dashboard → Deployments
2. Find previous working deployment
3. Click "Promote to Production"

### Backend
1. SSH into server
2. `cd /opt/rafiki`
3. `git log --oneline` (find commit to rollback to)
4. `git reset --hard <commit-hash>`
5. `/opt/rafiki/devops/deploy.sh`

## Monitoring

- UptimeRobot: https://uptimerobot.com
- Vercel Analytics: https://vercel.com/dashboard/analytics
- Grafana: https://grafana.rafiki.com

## Troubleshooting

### CORS Errors
- Verify backend CORS configuration: `docker compose exec rafiki-service env | grep CORS`
- Verify Nginx configuration: `docker compose exec nginx cat /etc/nginx/nginx.conf`
- Check browser console for error details

### API Not Responding
- Check backend health: `curl https://api.rafiki.com/v1/readiness`
- Check logs: `docker compose logs -f partner-service`
- Verify Nginx is running: `docker compose ps nginx`

### SSL Certificate Expired
- Renew certificate: `docker compose run --rm certbot renew`
- Reload Nginx: `docker compose exec nginx nginx -s reload`
```

### Acceptance Criteria

- [ ] Custom domain (app.rafiki.com) configured in Vercel
- [ ] DNS propagated and domain accessible
- [ ] SSL certificate valid
- [ ] Production CORS configured on backend
- [ ] Production deployment successful
- [ ] All production tests passing
- [ ] Monitoring set up (UptimeRobot)
- [ ] Documentation updated
- [ ] Team notified of production URL

---

## Testing Strategy

### Unit Testing (Optional for MVP)

```bash
# Install testing libraries
npm install --save-dev @testing-library/react @testing-library/jest-dom jest jest-environment-jsdom

# Create jest.config.js
# Add test files: *.test.tsx

# Example tests:
# - ThinkForm submission
# - API client functions
# - ThinkCard rendering
# - Pagination logic
```

### Integration Testing

**Manual testing checklist:**

- [ ] Create think with all 5 categories
- [ ] Create think with empty content (should fail)
- [ ] Create think with very long content (>1000 chars)
- [ ] View thinks list (empty state)
- [ ] View thinks list (with items)
- [ ] Pagination (previous/next buttons)
- [ ] Pagination (page numbers)
- [ ] Mobile responsive (iPhone SE, iPhone 12, iPad)
- [ ] Desktop responsive (1920x1080, 1366x768)
- [ ] Network error handling (disconnect internet)
- [ ] API error handling (500 error from backend)

### Load Testing

```bash
# Use Apache Bench or wrk
# Test API endpoint

# Test 1000 requests with 10 concurrent connections
ab -n 1000 -c 10 https://api.rafiki.com/v1/thinks

# Verify:
# - No 500 errors
# - Response times < 500ms
# - Rate limiting works (429 errors after threshold)
```

### Browser Compatibility

Test on:
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest)
- [ ] Edge (latest)
- [ ] Mobile Safari (iOS)
- [ ] Chrome Mobile (Android)

---

## Rollback Plan

### Scenario 1: Frontend Deployment Fails

**Steps:**

```bash
# Option 1: Vercel Dashboard
# 1. Go to Vercel Dashboard → Deployments
# 2. Find previous working deployment
# 3. Click "Promote to Production"

# Option 2: Vercel CLI
vercel rollback

# Option 3: Git Revert
git revert HEAD
git push origin main
# Vercel auto-deploys previous version
```

### Scenario 2: Backend Deployment Breaks CORS

**Steps:**

```bash
# SSH into server
ssh root@178.156.170.37

# Rollback to previous commit
cd /opt/rafiki
git log --oneline  # Find working commit
git reset --hard <commit-hash>

# Redeploy
/opt/rafiki/devops/deploy.sh

# Verify CORS
curl -i -X OPTIONS https://api.rafiki.com/v1/thinks \
  -H "Origin: https://app.rafiki.com" \
  -H "Access-Control-Request-Method: GET"
```

### Scenario 3: Database Issues

**Steps:**

```bash
# SSH into server
ssh root@178.156.170.37

# Check database status
docker compose ps postgres

# View logs
docker compose logs -f postgres

# If database is corrupted, restore from backup
# (Ensure you have regular backups!)
cat backup_YYYYMMDD_HHMMSS.sql | docker compose exec -T postgres psql -U rafiki rafiki

# Restart services
docker compose restart partner-service
```

---

## Post-Deployment Monitoring

### Week 1: Daily Checks

**Daily checklist:**

- [ ] Check UptimeRobot status (no downtime alerts)
- [ ] Check Vercel deployment status (no failed builds)
- [ ] Review Vercel Analytics (page views, errors)
- [ ] Check backend logs: `docker compose logs -f partner-service`
- [ ] Verify SSL certificates valid (expires in 90 days)
- [ ] Review Grafana dashboards for anomalies

### Week 2-4: Weekly Checks

**Weekly checklist:**

- [ ] Review UptimeRobot reports
- [ ] Review Vercel Analytics trends
- [ ] Check disk usage on Hetzner: `df -h`
- [ ] Check Docker disk usage: `docker system df`
- [ ] Review PostgreSQL performance
- [ ] Test backup restoration process
- [ ] Update dependencies if needed: `npm audit`

### Ongoing: Monthly Checks

**Monthly checklist:**

- [ ] Security updates on Hetzner server: `apt update && apt upgrade`
- [ ] Review and rotate logs
- [ ] Review SSL certificate expiration dates
- [ ] Performance optimization review
- [ ] Cost analysis (Vercel, Hetzner, domain)
- [ ] Backup verification
- [ ] Disaster recovery drill

### Metrics to Track

**Frontend (Vercel Analytics):**
- Page views
- Unique visitors
- Bounce rate
- Load time (p50, p75, p95)
- Error rate

**Backend (Grafana):**
- Request rate (requests/sec)
- Response time (p50, p75, p95, p99)
- Error rate (4xx, 5xx)
- Database query time
- CPU and memory usage

**Infrastructure:**
- Uptime percentage (target: 99.9%)
- SSL certificate validity
- Disk usage
- Docker container health

---

## Future Enhancements

### Phase 6: Authentication (Week 3-4)

**Features:**
- User registration and login
- JWT authentication
- Protected API endpoints
- User-specific thinks

**Tasks:**
- Implement JWT in Go backend
- Add authentication middleware
- Create login/signup pages in frontend
- Update API client with auth headers
- Migrate existing thinks to user-based model

### Phase 7: Enhanced Features (Month 2)

**Features:**
- Update/edit thinks
- Delete thinks
- Search and filter by category
- Tags system
- Markdown support in content
- Export to PDF/CSV

### Phase 8: Real-time Updates (Month 3)

**Features:**
- WebSocket support
- Real-time think creation notifications
- Collaborative editing
- Activity feed

### Phase 9: Mobile App (Month 4+)

**Features:**
- React Native mobile app
- Offline support
- Push notifications
- Native camera integration for image uploads

### Phase 10: Advanced Features (Month 6+)

**Features:**
- AI-powered think suggestions
- Sentiment analysis
- Think relationships (linking)
- Calendar integration
- Reminders and notifications
- Analytics dashboard

---

## Conclusion

This implementation plan provides a comprehensive, step-by-step guide to deploying the Rafiki "thinks" management frontend on Vercel, integrated with the existing Go backend on Hetzner.

### Key Success Factors

1. **Backend is ready:** All required API endpoints exist and are fully functional
2. **One critical fix:** CORS configuration needs to be wired to mux (2-minute fix)
3. **Clean architecture:** Nginx reverse proxy provides security, SSL, and rate limiting
4. **Modern stack:** Next.js 14, TypeScript, shadcn/ui, Tailwind CSS
5. **Production-ready:** Monitoring, rollback plan, comprehensive testing

### Timeline Summary

- **Phase 1 (Backend):** 1 hour
- **Phase 2 (Infrastructure):** 2-3 hours
- **Phase 3 (Frontend):** 6-8 hours
- **Phase 4 (Integration):** 2 hours
- **Phase 5 (Production):** 1 hour
- **Total:** 12-15 hours (1-2 days)

### Risk Mitigation

- **CORS issues:** Comprehensive CORS configuration with clear testing steps
- **SSL certificate:** Automated renewal with Certbot
- **Downtime:** Zero-downtime deployment strategy with health checks
- **Rollback:** Clear rollback procedures for all scenarios
- **Monitoring:** Proactive monitoring with UptimeRobot and Vercel Analytics

### Support & Resources

- **Backend Documentation:** [devops/DEPLOYMENT.md](../devops/DEPLOYMENT.md)
- **CLAUDE.md:** [CLAUDE.md](../CLAUDE.md)
- **Frontend Setup Prompt:** [frontend/SETUP_PROMPT.md](../frontend/SETUP_PROMPT.md)
- **Vercel Docs:** https://vercel.com/docs
- **shadcn/ui Docs:** https://ui.shadcn.com
- **Next.js Docs:** https://nextjs.org/docs

---

**Document Status:** Ready for Implementation
**Last Reviewed:** 2025-11-02
**Next Review:** After Phase 5 completion
