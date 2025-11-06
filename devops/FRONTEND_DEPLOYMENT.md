# Rafiki Frontend Deployment Guide - Vercel

## Overview

The Rafiki frontend is a Next.js 14+ application deployed on Vercel's global edge network. It connects to the Go backend API hosted on Hetzner servers.

## Architecture

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
│                    │          │                    │
│  Next.js Frontend  │──HTTPS───▶  Go Backend API   │
│  app.rafiki.com    │  Calls   │  api.rafiki.com    │
└────────────────────┘          └────────────────────┘
```

## Technology Stack

- **Framework:** Next.js 14+ (App Router)
- **Language:** TypeScript
- **UI Components:** shadcn/ui
- **Styling:** Tailwind CSS
- **Deployment:** Vercel
- **Domain:** app.rafiki.com (production)

## Prerequisites

- Node.js 18+ installed locally
- Vercel account (free tier sufficient for MVP)
- Git repository connected to Vercel
- Domain configured (optional for production)

## Project Structure

```
frontend/
├── app/                    # Next.js App Router pages
│   ├── layout.tsx         # Root layout with header
│   ├── page.tsx           # Home page
│   └── thinks/
│       └── page.tsx       # Thinks management page
├── components/
│   ├── ui/                # shadcn/ui components
│   ├── features/
│   │   ├── ThinkForm.tsx
│   │   ├── ThinkCard.tsx
│   │   └── ThinkList.tsx
│   └── layout/
│       └── Header.tsx
├── lib/
│   ├── api.ts             # API client
│   ├── types.ts           # TypeScript types
│   └── utils.ts           # Utility functions
├── public/                # Static assets
├── .env.local             # Local environment variables
├── .env.example           # Example environment variables
├── next.config.js         # Next.js configuration
├── tailwind.config.ts     # Tailwind configuration
└── package.json           # Dependencies
```

## Initial Setup

### 1. Install Dependencies

```bash
cd frontend
npm install
```

### 2. Configure Environment Variables

Create `.env.local` for local development:

```bash
# .env.local
NEXT_PUBLIC_API_URL=http://178.156.170.37:3000
NEXT_PUBLIC_ENV=development
```

For production (configured in Vercel):
```bash
NEXT_PUBLIC_API_URL=https://api.rafiki.com
NEXT_PUBLIC_ENV=production
```

### 3. Run Development Server

```bash
npm run dev
# Open http://localhost:3000
```

## Vercel Deployment

### Option 1: Vercel CLI (Recommended)

#### Install Vercel CLI
```bash
npm install -g vercel
```

#### Login to Vercel
```bash
vercel login
```

#### Link Project
```bash
cd frontend
vercel link
```

Follow the prompts:
- Set up and deploy? **Yes**
- Which scope? Select your account
- Link to existing project? **No** (first time) or **Yes** (subsequent)
- Project name? **rafiki-frontend**
- Code location? **./** (current directory)

#### Configure Environment Variables

```bash
# Production
vercel env add NEXT_PUBLIC_API_URL production
# Enter: https://api.rafiki.com

vercel env add NEXT_PUBLIC_ENV production
# Enter: production

# Preview
vercel env add NEXT_PUBLIC_API_URL preview
# Enter: https://api.rafiki.com

vercel env add NEXT_PUBLIC_ENV preview
# Enter: preview

# List environment variables
vercel env ls
```

#### Deploy to Preview

```bash
vercel
```

This deploys to a preview URL like: `rafiki-frontend-xxx.vercel.app`

#### Deploy to Production

```bash
vercel --prod
```

### Option 2: Vercel Dashboard (Git Integration)

#### Connect Git Repository

1. Go to [Vercel Dashboard](https://vercel.com/dashboard)
2. Click **"Add New Project"**
3. Import your Git repository (GitHub, GitLab, Bitbucket)
4. Select the repository
5. Configure project:
   - **Framework Preset:** Next.js
   - **Root Directory:** `frontend` (if monorepo) or `.` (if separate repo)
   - **Build Command:** `npm run build`
   - **Output Directory:** `.next`
   - **Install Command:** `npm install`

#### Configure Environment Variables

1. Go to **Settings → Environment Variables**
2. Add the following:

**Production:**
- `NEXT_PUBLIC_API_URL` = `https://api.rafiki.com`
- `NEXT_PUBLIC_ENV` = `production`

**Preview:**
- `NEXT_PUBLIC_API_URL` = `https://api.rafiki.com`
- `NEXT_PUBLIC_ENV` = `preview`

3. Click **Save**

#### Deploy

- **Automatic:** Push to main branch triggers production deployment
- **Manual:** Click **"Redeploy"** in Vercel Dashboard

## Custom Domain Configuration

### 1. Add Domain in Vercel

1. Go to **Project → Settings → Domains**
2. Click **"Add Domain"**
3. Enter: `app.rafiki.com`
4. Vercel will provide DNS instructions

### 2. Configure DNS

Add a CNAME record in your DNS provider:

```
Type:  CNAME
Name:  app
Value: cname.vercel-dns.com
TTL:   Auto or 3600
```

### 3. Verify Domain

1. Wait for DNS propagation (5-30 minutes)
2. Verify with: `dig app.rafiki.com`
3. In Vercel dashboard, click **"Refresh"** to verify
4. Vercel automatically provisions SSL certificate

## Backend CORS Configuration

The backend must allow requests from the frontend domain.

### On Hetzner Server

```bash
ssh root@178.156.170.37
cd /opt/rafiki
nano .env
```

Update CORS configuration:
```bash
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.com,https://rafiki-frontend-*.vercel.app,http://localhost:3000
```

Restart services:
```bash
docker compose down
docker compose up -d
```

### Verify CORS

```bash
curl -i -X OPTIONS https://api.rafiki.com/v1/thinks \
  -H "Origin: https://app.rafiki.com" \
  -H "Access-Control-Request-Method: GET"
```

Expected response headers:
```
Access-Control-Allow-Origin: https://app.rafiki.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

## Environment Variables Reference

### Frontend (Vercel)

| Variable | Development | Preview | Production | Description |
|----------|-------------|---------|------------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:3000` | `https://api.rafiki.com` | `https://api.rafiki.com` | Backend API base URL |
| `NEXT_PUBLIC_ENV` | `development` | `preview` | `production` | Environment identifier |

### Backend (Hetzner)

| Variable | Value | Description |
|----------|-------|-------------|
| `PARTNER_WEB_CORSALLOWEDORIGINS` | `https://app.rafiki.com,https://rafiki-frontend-*.vercel.app,http://localhost:3000` | Allowed CORS origins |

## Deployment Workflow

### Standard Deployment (Git Integration)

```bash
# 1. Make changes to frontend code
cd frontend
# Edit files...

# 2. Test locally
npm run dev

# 3. Commit and push
git add .
git commit -m "feat: add new feature"
git push origin main

# 4. Vercel automatically deploys to production
# Monitor at: https://vercel.com/dashboard
```

### Preview Deployment (Feature Branch)

```bash
# 1. Create feature branch
git checkout -b feature/new-feature

# 2. Make changes and push
git add .
git commit -m "feat: add new feature"
git push origin feature/new-feature

# 3. Vercel automatically creates preview deployment
# Preview URL: https://rafiki-frontend-git-feature-new-feature-xxx.vercel.app
```

### Manual Deployment (CLI)

```bash
cd frontend

# Deploy to preview
vercel

# Deploy to production
vercel --prod

# Rollback to previous deployment
vercel rollback
```

## Monitoring & Analytics

### Vercel Analytics (Built-in)

1. Go to **Project → Analytics**
2. View metrics:
   - Page views
   - Unique visitors
   - Top pages
   - Traffic sources
   - Performance metrics (Web Vitals)

### Custom Analytics (Optional)

Install Vercel Analytics package:

```bash
npm install @vercel/analytics
```

Update `app/layout.tsx`:

```typescript
import { Analytics } from '@vercel/analytics/react';

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  );
}
```

### External Monitoring

Set up monitoring with:
- **UptimeRobot:** Monitor `https://app.rafiki.com` availability
- **Sentry:** Error tracking (optional)
- **LogRocket:** Session replay (optional)

## Troubleshooting

### Build Fails on Vercel

**Check build logs:**
1. Go to **Deployments** tab
2. Click on failed deployment
3. View build logs

**Common issues:**
- Missing environment variables
- TypeScript errors
- Build command incorrect
- Node version mismatch

**Solution:**
```bash
# Test build locally
npm run build

# Check Node version
node --version  # Should be 18+

# Fix TypeScript errors
npm run type-check
```

### CORS Errors

**Symptoms:**
- "Access to fetch blocked by CORS policy" in browser console
- OPTIONS preflight requests fail

**Diagnosis:**
```bash
# Check CORS configuration on backend
curl -i -X OPTIONS https://api.rafiki.com/v1/thinks \
  -H "Origin: https://app.rafiki.com" \
  -H "Access-Control-Request-Method: GET"
```

**Solutions:**
1. Verify backend CORS configuration includes frontend domain
2. Restart backend services after CORS changes
3. Check that backend is accessible from internet
4. Verify SSL certificates are valid

### API Not Responding

**Symptoms:**
- Network errors in browser console
- Timeout errors
- 502/503 errors

**Diagnosis:**
```bash
# Check backend health
curl https://api.rafiki.com/v1/readiness
curl https://api.rafiki.com/v1/liveness

# Check backend logs
ssh root@178.156.170.37
docker compose logs -f partner-service
```

**Solutions:**
1. Verify backend is running: `docker compose ps`
2. Check backend logs for errors
3. Verify firewall allows traffic on port 3000
4. Verify Nginx/reverse proxy configuration (if applicable)

### Environment Variables Not Working

**Symptoms:**
- `process.env.NEXT_PUBLIC_API_URL` is undefined
- API calls go to wrong URL

**Diagnosis:**
```bash
# Check environment variables in Vercel
vercel env ls

# Check local environment variables
cat .env.local
```

**Solutions:**
1. Ensure variables start with `NEXT_PUBLIC_` prefix
2. Redeploy after adding environment variables
3. Check that variables are set for correct environment (production/preview/development)

### Preview Deployments Not Working

**Symptoms:**
- Preview URL returns 404
- Preview deployment fails

**Diagnosis:**
1. Check deployment status in Vercel dashboard
2. View deployment logs

**Solutions:**
1. Verify branch name doesn't contain invalid characters
2. Ensure build command succeeds
3. Check that preview environment variables are set

## Performance Optimization

### 1. Image Optimization

Use Next.js Image component:

```typescript
import Image from 'next/image';

<Image
  src="/logo.png"
  alt="Logo"
  width={200}
  height={200}
  priority  // For above-the-fold images
/>
```

### 2. Code Splitting

Next.js automatically code-splits by route. For component-level splitting:

```typescript
import dynamic from 'next/dynamic';

const HeavyComponent = dynamic(() => import('./HeavyComponent'), {
  loading: () => <p>Loading...</p>,
  ssr: false  // Disable server-side rendering if not needed
});
```

### 3. API Response Caching

Configure Next.js fetch caching:

```typescript
// Cache for 60 seconds
const data = await fetch('https://api.rafiki.com/v1/thinks', {
  next: { revalidate: 60 }
});
```

### 4. Edge Functions (Optional)

For server-side logic, use Edge Functions:

```typescript
// app/api/hello/route.ts
export const runtime = 'edge';

export async function GET() {
  return new Response('Hello from Edge!');
}
```

## Security Best Practices

### 1. Environment Variables

- ✅ Use `NEXT_PUBLIC_` prefix for client-side variables
- ✅ Never commit `.env.local` to version control
- ✅ Use Vercel's environment variable UI for secrets
- ❌ Don't expose sensitive data in `NEXT_PUBLIC_` variables

### 2. API Security

- ✅ Validate all user input on frontend AND backend
- ✅ Use HTTPS for all API calls
- ✅ Implement rate limiting on backend
- ✅ Use Content Security Policy (CSP) headers

### 3. Dependencies

```bash
# Regular security audits
npm audit

# Update dependencies
npm update

# Check for outdated packages
npm outdated
```

## Rollback Strategy

### Option 1: Vercel Dashboard

1. Go to **Deployments** tab
2. Find previous working deployment
3. Click **three dots (...)** → **Promote to Production**

### Option 2: Vercel CLI

```bash
vercel rollback
```

### Option 3: Git Revert

```bash
# Revert last commit
git revert HEAD
git push origin main

# Vercel auto-deploys previous version
```

## CI/CD Integration

### GitHub Actions (Optional)

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Vercel

on:
  push:
    branches: [main]
    paths:
      - 'frontend/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install dependencies
        run: cd frontend && npm ci

      - name: Run tests
        run: cd frontend && npm test

      - name: Build
        run: cd frontend && npm run build

      - name: Deploy to Vercel
        run: cd frontend && vercel --prod --token=${{ secrets.VERCEL_TOKEN }}
```

## Cost Estimation

### Vercel Pricing (as of 2024)

**Free (Hobby) Plan:**
- ✅ 100 GB bandwidth/month
- ✅ 100 GB-hours compute/month
- ✅ Unlimited deployments
- ✅ SSL certificates
- ✅ Analytics (basic)
- ❌ No team collaboration
- ❌ No password protection

**Pro Plan ($20/month):**
- ✅ 1 TB bandwidth/month
- ✅ 1000 GB-hours compute/month
- ✅ Team collaboration
- ✅ Password protection
- ✅ Advanced analytics
- ✅ Priority support

**Recommendation:** Start with Free plan for MVP, upgrade to Pro when needed.

## Maintenance Checklist

### Daily
- [ ] Check Vercel deployment status
- [ ] Monitor analytics for errors
- [ ] Review Vercel function logs

### Weekly
- [ ] Check for dependency updates: `npm outdated`
- [ ] Review performance metrics (Web Vitals)
- [ ] Test critical user flows

### Monthly
- [ ] Security audit: `npm audit`
- [ ] Update dependencies: `npm update`
- [ ] Review Vercel usage and costs
- [ ] Test disaster recovery (rollback)

## Additional Resources

- **Next.js Documentation:** https://nextjs.org/docs
- **Vercel Documentation:** https://vercel.com/docs
- **shadcn/ui Documentation:** https://ui.shadcn.com
- **Tailwind CSS Documentation:** https://tailwindcss.com/docs
- **Backend Deployment Guide:** [DEPLOYMENT.md](./DEPLOYMENT.md)
- **Frontend Implementation Plan:** [../docs/FRONTEND_IMPLEMENTATION_PLAN.md](../docs/FRONTEND_IMPLEMENTATION_PLAN.md)

## Support

For issues or questions:
1. Check Vercel status page: https://vercel-status.com
2. Review Vercel documentation
3. Check backend health: [DEPLOYMENT.md](./DEPLOYMENT.md)
4. Review frontend logs in Vercel dashboard
5. Check CORS configuration on backend

---

**Last Updated:** 2025-11-03
**Maintained By:** Rafiki Development Team
**Next Review:** After production deployment
