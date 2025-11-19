# Frontend Development Guide

This guide covers frontend development workflows, code quality tools, and best practices for the Rafiki Next.js application.

## Tech Stack

- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript 5
- **UI Library**: React 19
- **Styling**: Tailwind CSS 4
- **Forms**: React Hook Form + Zod
- **UI Components**: Radix UI
- **Deployment**: Vercel

## Quick Start

### Prerequisites

- Node.js 20+ installed
- npm or pnpm package manager

### Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Open http://localhost:3000
```

## Code Quality Tools

The frontend uses automated code quality tools as part of the CodeRabbit automation setup (Phase 1).

### Prettier (Formatting)

**Configuration**: [.prettierrc](../frontend/.prettierrc)

- Single quotes for JS/TS, double quotes for JSX
- 100 character line width
- 2-space indentation
- Always use semicolons
- ES5 trailing commas
- LF line endings

**Commands**:
```bash
# Check formatting (CI)
npm run format:check

# Apply formatting (local dev)
npm run format

# Auto-formats on commit (via pre-commit hooks in Phase 3)
```

### ESLint (Linting)

**Configuration**: [eslint.config.mjs](../frontend/eslint.config.mjs)

**Auto-fixable rules (Tier 1)**:
- Code style: quotes, semicolons, trailing commas
- React patterns: self-closing components, boolean attributes
- Import organization

**Warning rules (Tier 2)**:
- `@typescript-eslint/no-explicit-any`: Warn about `any` types
- `no-console`: Warn about console.log (allow warn/error)
- React hooks dependencies

**Error rules (Tier 3)**:
- `no-debugger`: No debugger statements
- `prefer-const`: Use const when possible
- `no-var`: Never use var
- Unused variables (except prefixed with `_`)

**Commands**:
```bash
# Lint code
npm run lint

# Auto-fix safe issues
npm run lint:fix

# Runs automatically in CI and pre-commit hooks
```

### TypeScript

**Configuration**: [tsconfig.json](../frontend/tsconfig.json)

**Commands**:
```bash
# Type check (no output = success)
npm run typecheck

# Runs automatically during build
npm run build
```

### Combined Checks

```bash
# Run all checks (typecheck + lint + format check)
npm run check
```

## Development Workflow

### Branch Strategy

**IMPORTANT**: The `main` branch is protected. All changes must go through pull requests.

1. **Create feature branch**:
   ```bash
   git checkout -b feature/my-feature
   # or: fix/bug-name, chore/task-name
   ```

2. **Make changes and commit**:
   ```bash
   git add .
   git commit -m "feat: add user profile page"
   # Pre-commit hooks auto-format code (Phase 3)
   ```

3. **Push and create PR**:
   ```bash
   git push origin feature/my-feature
   gh pr create --title "Add user profile page" --body "Description"
   ```

4. **CodeRabbit review**:
   - CodeRabbit AI reviews PR automatically
   - Provides feedback categorized by tier (auto-fix, manual, architectural)
   - Use `/coderabbit-review` command to re-trigger (Phase 4)

5. **Address feedback**:
   - Fix any issues found
   - Push updates to same branch
   - CodeRabbit re-reviews automatically

6. **Get approval & merge**:
   - Requires 1 approval from team member
   - All CI checks must pass (Phase 2)
   - Merge creates linear history (no merge commits)

### Commit Message Format

Use conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `chore`: Maintenance (configs, deps)
- `docs`: Documentation
- `style`: Formatting (no logic change)
- `refactor`: Code restructuring
- `test`: Adding tests
- `perf`: Performance improvement

**Examples**:
```bash
git commit -m "feat(auth): add password reset flow"
git commit -m "fix(moments): correct intensity validation"
git commit -m "chore(deps): update Next.js to 16.0.2"
```

## Code Patterns

### Project Structure

```
frontend/
├── app/                      # Next.js App Router
│   ├── (dashboard)/         # Route group (layout)
│   │   ├── layout.tsx       # Dashboard layout
│   │   ├── page.tsx         # Dashboard home
│   │   ├── momentos/        # Moments feature
│   │   └── thinks/          # Thinks feature
│   ├── login/               # Public route
│   ├── layout.tsx           # Root layout
│   └── globals.css          # Global styles
├── components/
│   ├── auth/                # Authentication components
│   ├── dashboard/           # Dashboard components
│   ├── features/            # Feature-specific components
│   ├── layout/              # Layout components
│   └── ui/                  # Reusable UI components (shadcn/ui)
├── lib/
│   ├── api.ts               # API client
│   ├── auth-context.tsx     # Auth context provider
│   ├── types.ts             # TypeScript types
│   └── utils.ts             # Utility functions
└── public/                  # Static assets
```

### Component Patterns

**Server Components (default)**:
```typescript
// app/momentos/page.tsx
export default async function MomentosPage() {
  // Fetch data on server
  const moments = await fetchMoments();

  return (
    <div>
      <MomentList moments={moments} />
    </div>
  );
}
```

**Client Components** (use sparingly):
```typescript
'use client';

import { useState } from 'react';

export function MomentForm() {
  const [intensity, setIntensity] = useState(5);

  // Interactive logic here

  return <form>...</form>;
}
```

### API Client Pattern

**Location**: `lib/api.ts`

```typescript
// Type-safe API calls
const response = await apiClient.post('/v1/moments', {
  situation: 'Working on feature',
  intensity: 7,
});

// Error handling
try {
  await apiClient.delete(`/v1/moments/${id}`);
} catch (error) {
  console.error('Failed to delete moment:', error);
}
```

### Authentication

**Context**: `lib/auth-context.tsx`

```typescript
'use client';

import { useAuth } from '@/lib/auth-context';

export function UserProfile() {
  const { user, loading, logout } = useAuth();

  if (loading) return <LoadingSpinner />;
  if (!user) return <LoginPrompt />;

  return <div>Welcome {user.name}</div>;
}
```

### Form Validation with Zod

```typescript
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';

const momentSchema = z.object({
  situation: z.string().min(1, 'Situation is required'),
  intensity: z.number().min(0).max(10),
});

export function MomentForm() {
  const form = useForm({
    resolver: zodResolver(momentSchema),
  });

  // ...
}
```

## Testing

### Unit Tests (TODO - Phase 3)
```bash
npm run test
npm run test:watch
npm run test:coverage
```

### E2E Tests (TODO - Phase 3)
```bash
npm run test:e2e
```

## Building for Production

```bash
# Create optimized production build
npm run build

# Test production build locally
npm run start

# Build output: .next/
```

## Environment Variables

Create `.env` file in frontend directory:

```bash
# API endpoint
NEXT_PUBLIC_API_URL=http://localhost:3000

# Optional: Enable debug logging
NEXT_PUBLIC_DEBUG=true
```

**IMPORTANT**: Never commit `.env` files. Use `.env.example` for templates.

## Common Issues

### Build Errors

**Issue**: `Type error: Cannot find module`
```bash
# Solution: Check imports and type definitions
npm run typecheck
```

**Issue**: `Module not found: Can't resolve '@/components/...`
```bash
# Solution: Ensure tsconfig paths are correct
cat tsconfig.json | grep paths
```

### Styling Issues

**Issue**: Tailwind classes not working
```bash
# Solution: Ensure Tailwind is watching files
npm run dev
# Check tailwind.config.ts content paths
```

### Hydration Errors

**Issue**: `Text content does not match server-rendered HTML`
```typescript
// Solution: Avoid client-only data in server components
// Use 'use client' directive or useEffect
```

## CodeRabbit Integration

### Path-Specific Rules

CodeRabbit has special instructions for frontend paths:

- **`frontend/lib/auth-context.tsx`**: Security-sensitive, never auto-fix
- **`frontend/lib/api.ts`**: API contract, manual review required
- **`frontend/app/**`**: Next.js 16 App Router conventions

### Review Tiers

**Tier 1 (Auto-fix)**: CodeRabbit auto-fixes and commits
- Formatting issues (Prettier)
- Simple style issues (ESLint auto-fixable)
- Import organization

**Tier 2 (Interactive)**: CodeRabbit suggests, you approve
- Type improvements (`any` → specific types)
- React best practices
- Performance optimizations

**Tier 3 (Manual)**: Flagged for manual review
- API contract changes
- Authentication logic
- Security-sensitive code

## Performance Best Practices

### Code Splitting
```typescript
// Lazy load heavy components
import dynamic from 'next/dynamic';

const HeavyChart = dynamic(() => import('./HeavyChart'), {
  loading: () => <Skeleton />,
  ssr: false, // Disable SSR if needed
});
```

### Image Optimization
```typescript
import Image from 'next/image';

<Image
  src="/profile.jpg"
  alt="Profile"
  width={200}
  height={200}
  priority // For above-the-fold images
/>
```

### Metadata for SEO
```typescript
// app/page.tsx
export const metadata = {
  title: 'Dashboard - Rafiki',
  description: 'Track your ideals, values, and habits',
};
```

## Deployment

Frontend is deployed to **Vercel** automatically on push to `main`.

### Vercel Configuration

**File**: [vercel.json](../frontend/vercel.json)

```json
{
  "buildCommand": "npm run build",
  "outputDirectory": ".next",
  "framework": "nextjs",
  "rewrites": [
    {
      "source": "/api/:path*",
      "destination": "https://api.rafiki.lat/:path*"
    }
  ]
}
```

### Deployment URLs

- **Production**: https://app.rafiki.lat
- **Preview**: Auto-deployed for each PR
- **API Backend**: https://api.rafiki.lat

### Manual Deploy

```bash
# Install Vercel CLI
npm install -g vercel

# Deploy
vercel deploy

# Deploy to production
vercel deploy --prod
```

## Getting Help

- **CodeRabbit issues**: Check `.coderabbit.yaml` configuration
- **Build errors**: Run `npm run check` for all validations
- **TypeScript errors**: Run `npm run typecheck`
- **Styling issues**: Verify Tailwind configuration

## Related Documentation

- [Backend Development Guide](./BACKEND_DEVELOPMENT.md)
- [Deployment Guide](./DEPLOYMENT_GUIDE.md)
- [Frontend Deployment Guide](./FRONTEND_DEPLOYMENT.md)
- [CodeRabbit Configuration](../docs/coderabbit-automation-implementation-plan.md)
