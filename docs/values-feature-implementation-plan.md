# Values Feature - Complete Implementation Plan

**Project:** Rafiki Habits Tracker
**Feature:** Personal Values Tracking with Facet Categorization
**Date:** November 22, 2025
**Status:** Ready for Implementation

---

## Executive Summary

This document provides a complete implementation plan for adding a "values" feature to the Rafiki habits tracking application. The feature allows users to define and track their core personal values (maximum 10) with life facet categorization.

### Key Requirements Confirmed

**User Requirements:**
- ✅ Values have life facet tags/categories (ENUM with 8 predefined values)
- ✅ Maximum 200 characters per value
- ✅ Include creation/modification dates
- ✅ User-controlled display_order (priority ranking)
- ✅ Limit: 10 values max per user (with UI explanation)
- ✅ Validation: 3-200 characters
- ✅ Plain text, encrypted content
- ✅ Private (user-scoped, not shared)
- ✅ Simple integer-based reordering
- ✅ Field-level encryption

**Design Requirements:**
- ✅ Card-based display
- ✅ Dedicated `/values` page + preview on main dashboard
- ✅ Show all values (no pagination needed for max 10)
- ✅ Visual hierarchy: first value most prominent
- ✅ Scroll-to-top form for editing
- ✅ Active state only (no archived states)
- ✅ Rose/crimson color scheme (expert choice)
- ✅ No emoji support initially
- ✅ Always-visible creation form
- ✅ Mobile-first responsive design

**Technical Requirements:**
- ✅ Fetch values on page mount
- ✅ useState for state management
- ✅ Explicit save button (no auto-save)
- ✅ Inline error handling
- ✅ Server confirmation before UI updates
- ✅ Simple refetch pattern (no caching)
- ✅ No unsaved changes warnings (v1)
- ✅ No real-time updates
- ✅ No keyboard shortcuts (v1)
- ✅ Responsive grid + Sheet forms

**Deployment Requirements:**
- ✅ No feature flag
- ✅ MVP deployment (flexible timing)
- ✅ Sequential deployment (backend first, then frontend)
- ✅ Automatic migration during deployment
- ✅ Users start with empty values
- ✅ Basic logs + health checks
- ✅ Deploy and monitor (no load testing)
- ✅ Rely on automatic PlanetScale backups

---

## Architecture Overview

### Database Layer

**New ENUM Type:**
```sql
CREATE TYPE facet_type AS ENUM (
    'health',
    'relationships',
    'career',
    'personal_growth',
    'family',
    'creativity',
    'community',
    'spirituality',
    'leisure',
    'financial'
);
```

**New Table:**
```sql
CREATE TABLE values (
    value_id       UUID        NOT NULL,
    user_id        UUID        NOT NULL,
    content        TEXT        NOT NULL,  -- Encrypted
    facet          facet_type  NOT NULL,
    display_order  INTEGER     NOT NULL,
    date_created   TIMESTAMP   NOT NULL,
    date_updated   TIMESTAMP   NOT NULL,
    PRIMARY KEY (value_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

**Indexes:**
- `values_user_id_idx` - Query by user
- `values_user_order_idx` - Sort by display_order
- `values_facet_idx` - Filter by facet
- `values_user_order_unique_idx` - Prevent duplicate priorities

### Backend API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/values` | Create new value (enforces 10 max) |
| GET | `/v1/values` | List user's values (sorted by display_order) |
| GET | `/v1/values?facet={facet}` | Filter values by facet |
| GET | `/v1/values/{id}` | Get single value by ID |
| PUT | `/v1/values/{id}` | Update value |
| DELETE | `/v1/values/{id}` | Delete value |

### Frontend Pages

1. **Dashboard Integration** (`/dashboard`)
   - Values preview widget (top 3-5 values)
   - QuickStatsCard showing count
   - FeatureCard linking to `/values`

2. **Values Page** (`/values`)
   - Page header with explanation
   - 10-value limit notice
   - Always-visible creation form
   - Grid of value cards (1/2/3 columns responsive)

### Life Facets (8 Categories)

Based on ACT therapy and life domains research:

1. **Health** - Physical and mental wellbeing
2. **Relationships** - Family, friends, connections
3. **Career** - Professional development, work
4. **Personal Growth** - Self-improvement, learning, spirituality
5. **Family** - Family relationships specifically
6. **Creativity** - Creative expression, arts
7. **Community** - Service, contribution, helping others
8. **Spirituality** - Meaning, purpose, transcendence

---

## Implementation Checklist

### Phase 1: Backend Implementation

**Database Migration (Version 1.04):**
- [ ] Add migration SQL to `business/sdk/migrate/sql/migrate.sql`
- [ ] Test migration locally
- [ ] Verify ENUM type creation
- [ ] Verify table structure and indexes

**Business Types:**
- [ ] Create `business/types/facet/facet.go` (ENUM-like type)
- [ ] Create `business/types/valuecontent/valuecontent.go` (or reuse content.Content)
- [ ] Add validation (3-200 chars)

**Domain Layer:**
- [ ] Create `business/domain/valuebus/model.go`
- [ ] Create `business/domain/valuebus/valuebus.go` (business logic)
- [ ] Create `business/domain/valuebus/filter.go`
- [ ] Create `business/domain/valuebus/order.go`
- [ ] Implement 10-value limit enforcement

**Store Layer:**
- [ ] Create `business/domain/valuebus/stores/valuedb/model.go`
- [ ] Create `business/domain/valuebus/stores/valuedb/valuedb.go`
- [ ] Create `business/domain/valuebus/stores/valuedb/order.go`
- [ ] Implement encryption/decryption at boundary

**App Layer:**
- [ ] Create `app/domain/valueapp/valueapp.go` (HTTP handlers)
- [ ] Create `app/domain/valueapp/model.go` (request/response DTOs)
- [ ] Create `app/domain/valueapp/route.go`
- [ ] Create `app/domain/valueapp/order.go`

**Integration:**
- [ ] Update `api/services/partners/main.go` (register ValueBus)
- [ ] Update `app/sdk/mux/mux.go` (add ValueBus to BusConfig)
- [ ] Update `api/services/partners/all/all.go` (register routes)

**Testing:**
- [ ] Test all 5 endpoints locally
- [ ] Verify encryption works
- [ ] Test max 10 values limit
- [ ] Test facet filtering
- [ ] Run `golangci-lint run --fix`

### Phase 2: Frontend Implementation

**TypeScript Types:**
- [ ] Add `Facet` type to `lib/types.ts`
- [ ] Add `Value` interface to `lib/types.ts`
- [ ] Add `NewValue` interface to `lib/types.ts`
- [ ] Add `UpdateValue` interface to `lib/types.ts`
- [ ] Add `ValueListResponse` interface to `lib/types.ts`

**API Client:**
- [ ] Add `api.values.getAll()` to `lib/api.ts`
- [ ] Add `api.values.getById()` to `lib/api.ts`
- [ ] Add `api.values.create()` to `lib/api.ts`
- [ ] Add `api.values.update()` to `lib/api.ts`
- [ ] Add `api.values.delete()` to `lib/api.ts`

**Helper Utilities:**
- [ ] Create `lib/value-utils.ts` (facet colors, icons, labels)

**Components:**
- [ ] Create `components/features/ValueCard.tsx`
- [ ] Create `components/features/ValueForm.tsx` (React Hook Form + Zod)
- [ ] Create `components/features/ValueList.tsx`
- [ ] Create `components/dashboard/ValuesPreview.tsx` (optional)

**Pages:**
- [ ] Create `app/(dashboard)/values/page.tsx`
- [ ] Update `app/(dashboard)/page.tsx` (enable FeatureCard, add preview)

**Testing:**
- [ ] Test form validation
- [ ] Test CRUD operations
- [ ] Test responsive breakpoints
- [ ] Test accessibility (keyboard nav)
- [ ] Run `npm run check` (lint + format + typecheck)

### Phase 3: Deployment

**Pre-Deployment:**
- [ ] Create feature branch (`feature/values-mvp`)
- [ ] Commit all changes with clear message
- [ ] Push and create PR
- [ ] Approve PR (can self-approve)
- [ ] Merge to main

**Backend Deployment:**
- [ ] Run `make deploy` (deploys to Hetzner)
- [ ] Wait for health checks (max 60 seconds)
- [ ] Verify migration 1.04 applied (check logs)
- [ ] Test all 5 endpoints on production
- [ ] Monitor logs for errors

**Frontend Deployment:**
- [ ] Run `vercel --prod` OR merge to main (auto-deploy)
- [ ] Wait for build (2-3 minutes)
- [ ] Test `/values` page loads
- [ ] Test values CRUD in browser

**Verification:**
- [ ] Health checks pass
- [ ] All API endpoints work
- [ ] Frontend pages load
- [ ] Database table exists
- [ ] ENUM type created
- [ ] No errors in logs

### Phase 4: Post-Deployment Monitoring

**First Hour:**
- [ ] Monitor logs every 15 minutes
- [ ] Health check every 15 minutes
- [ ] Test creating a value as real user

**First Day:**
- [ ] Health check every hour
- [ ] Review error logs every 4 hours
- [ ] Monitor database value count

**First Week:**
- [ ] Daily health checks
- [ ] Weekly error log review
- [ ] Collect user feedback (if announced)

---

## File Structure

### Backend Files (Go)

```
business/
├── types/
│   ├── facet/
│   │   └── facet.go                      # NEW: Facet ENUM type
│   └── valuecontent/
│       └── valuecontent.go               # NEW: Value content validation
├── domain/
│   └── valuebus/
│       ├── model.go                      # NEW: Domain models
│       ├── valuebus.go                   # NEW: Business logic
│       ├── filter.go                     # NEW: Query filtering
│       ├── order.go                      # NEW: Ordering
│       └── stores/
│           └── valuedb/
│               ├── model.go              # NEW: DB models
│               ├── valuedb.go            # NEW: DB operations
│               └── order.go              # NEW: DB ordering
└── sdk/
    └── migrate/
        └── sql/
            └── migrate.sql               # MODIFIED: Add version 1.04

app/
└── domain/
    └── valueapp/
        ├── valueapp.go                   # NEW: HTTP handlers
        ├── model.go                      # NEW: Request/response DTOs
        ├── route.go                      # NEW: Route registration
        └── order.go                      # NEW: Order fields

api/services/partners/
├── main.go                               # MODIFIED: Register ValueBus
├── all/
│   └── all.go                            # MODIFIED: Register routes
└── mux/
    └── mux.go                            # MODIFIED: Add ValueBus to BusConfig
```

### Frontend Files (TypeScript/React)

```
frontend/
├── lib/
│   ├── types.ts                          # MODIFIED: Add Value types
│   ├── api.ts                            # MODIFIED: Add values API
│   └── value-utils.ts                    # NEW: Helper functions
├── components/
│   ├── features/
│   │   ├── ValueCard.tsx                 # NEW: Value display card
│   │   ├── ValueForm.tsx                 # NEW: Create/edit form
│   │   └── ValueList.tsx                 # NEW: List container
│   └── dashboard/
│       └── ValuesPreview.tsx             # NEW: Dashboard widget
└── app/
    └── (dashboard)/
        ├── page.tsx                      # MODIFIED: Add values preview
        └── values/
            └── page.tsx                  # NEW: Main values page
```

---

## API Examples

### Create Value

```bash
POST /v1/values
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Live with integrity and authenticity",
  "facet": "personal_growth",
  "display_order": 1
}

# Response (201 Created):
{
  "id": "uuid",
  "content": "Live with integrity and authenticity",
  "facet": "personal_growth",
  "displayOrder": 1,
  "dateCreated": "2025-11-22T10:00:00Z",
  "dateUpdated": "2025-11-22T10:00:00Z"
}
```

### List Values

```bash
GET /v1/values
Authorization: Bearer <token>

# Response (200 OK):
{
  "items": [
    {
      "id": "uuid",
      "content": "Live with integrity...",
      "facet": "personal_growth",
      "displayOrder": 1,
      "dateCreated": "2025-11-22T10:00:00Z",
      "dateUpdated": "2025-11-22T10:00:00Z"
    }
  ],
  "total": 1
}
```

### Filter by Facet

```bash
GET /v1/values?facet=personal_growth
Authorization: Bearer <token>

# Returns only values with facet="personal_growth"
```

### Update Value

```bash
PUT /v1/values/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "UPDATED: Live with complete integrity",
  "displayOrder": 2
}

# Response (200 OK): Updated value object
```

### Delete Value

```bash
DELETE /v1/values/{id}
Authorization: Bearer <token>

# Response (204 No Content)
```

---

## Design Specifications

### Color Scheme

**Primary Feature Color:** Rose/Crimson (#E11D48 / rose-600)
- Warmer and more humanistic than pure red
- Better conveys personal connection and introspection
- WCAG AA compliant

**Facet Colors:**
- Health: Emerald (#059669)
- Relationships: Blue (#2563EB)
- Career: Amber (#D97706)
- Personal Growth: Purple (#9333EA)
- Family: Pink (#EC4899)
- Creativity: Orange (#F97316)
- Community: Green (#16A34A)
- Spirituality: Indigo (#4F46E5)

### Visual Hierarchy

**#1 Value (Most Important):**
- Ring border (2px rose-500)
- Gradient background (rose-50 to white)
- Badge: "#1 Core Value"
- Larger text size

**Other Values:**
- Standard card design
- Numbered priority badge (#2, #3, etc.)
- Facet badge with icon and color

### Responsive Breakpoints

| Screen Size | Columns | Layout |
|-------------|---------|--------|
| < 640px (Mobile) | 1 | Stack vertically |
| 640-1024px (Tablet) | 2 | Side-by-side |
| > 1024px (Desktop) | 3 | Grid |

### Accessibility

- WCAG AA compliant colors (4.5:1 contrast)
- Keyboard navigable (Tab order, Enter/Escape)
- ARIA labels on all interactive elements
- Screen reader friendly
- Focus indicators visible
- Touch targets min 44px

---

## Deployment Timeline

**Estimated Total Time:** 55 minutes

| Phase | Duration |
|-------|----------|
| Pre-deployment checks | 5 min |
| Local testing | 10 min |
| Git PR workflow | 5 min |
| Backend deployment | 6 min |
| Backend verification | 5 min |
| API smoke tests | 5 min |
| Frontend deployment | 8 min |
| Frontend verification | 5 min |
| Database verification | 3 min |
| Monitoring setup | 3 min |
| **TOTAL** | **55 min** |

**Recommended Timing:** Weekday morning (10 AM - 12 PM)

---

## Rollback Strategy

**Backend Rollback:**
1. SSH to server
2. `git reset --hard <previous-commit>`
3. `sudo ./devops/deploy.sh`
4. **Recovery Time:** 3-5 minutes

**Frontend Rollback:**
1. Vercel Dashboard → Deployments
2. Select previous deployment
3. Click "Promote to Production"
4. **Recovery Time:** 30 seconds

**Database Rollback (if needed):**
```sql
DROP TABLE IF EXISTS values CASCADE;
DROP TYPE IF EXISTS facet_type;
DELETE FROM darwin_migrations WHERE version = '1.04';
```
**Recovery Time:** 2-3 minutes
**Data Loss:** All user values (acceptable for MVP rollback)

---

## Success Criteria

**Deployment Success:**
- ✅ Health checks pass (`/v1/readiness`)
- ✅ All 5 API endpoints work on production
- ✅ Frontend `/values` page loads
- ✅ Database migration 1.04 applied
- ✅ ENUM type `facet_type` created
- ✅ Values table exists with correct structure
- ✅ No errors in logs

**Feature Success (First Week):**
- ✅ Users can create values (max 10)
- ✅ Users can edit and delete values
- ✅ Facet filtering works
- ✅ Display ordering works
- ✅ Encryption verified (content not plain text in DB)
- ✅ No 500 errors
- ✅ Mobile responsive working

---

## Monitoring & Support

**Health Checks:**
```bash
# Readiness
curl https://api.rafiki.lat/v1/readiness

# Liveness
curl https://api.rafiki.lat/v1/liveness
```

**Log Monitoring:**
```bash
# Real-time logs
make deploy-logs

# Filter for values
ssh root@178.156.170.37 'docker compose logs -f partner-service | grep -i "value"'

# Error monitoring
ssh root@178.156.170.37 'docker compose logs -f partner-service | grep -i "error"'
```

**Database Queries:**
```bash
# Check migration applied
psql -h us-east-5.pg.psdb.cloud ... -c \
  "SELECT version FROM darwin_migrations WHERE version = '1.04';"

# Count values
psql -h us-east-5.pg.psdb.cloud ... -c \
  "SELECT COUNT(*) FROM values;"

# Check ENUM type
psql -h us-east-5.pg.psdb.cloud ... -c \
  "SELECT enumlabel FROM pg_enum WHERE enumtypid = 'facet_type'::regtype;"
```

---

## Troubleshooting

**Issue: POST /v1/values returns 500**
- Check encryption key in `.env`
- Verify database connection
- Check logs for specific error

**Issue: Frontend shows "Failed to load values"**
- Check CORS settings
- Test endpoint with curl
- Check authentication token

**Issue: "Maximum 10 values allowed" when < 10**
- Query database for actual count
- Check business logic in valuebus.go

**Issue: Facet filter returns no results**
- Verify facet value matches ENUM exactly (lowercase, underscore)
- Check SQL query in logs

---

## Next Steps

1. **Implementation:** Follow Phase 1-4 checklist above
2. **Code Review:** Use CodeRabbit for PR review
3. **Testing:** Local testing before deployment
4. **Deployment:** Use deployment runbook (see deployment plan)
5. **Monitoring:** Track health for first 24 hours
6. **Iteration:** Collect user feedback and iterate

---

## Documentation References

- **Backend Development:** `/Users/francowini/Documents/rafiki/devops/BACKEND_DEVELOPMENT.md`
- **Frontend Development:** `/Users/francowini/Documents/rafiki/devops/FRONTEND_DEVELOPMENT.md`
- **Deployment Guide:** `/Users/francowini/Documents/rafiki/devops/DEPLOYMENT_GUIDE.md`
- **This Implementation Plan:** `/Users/francowini/Documents/rafiki/docs/values-feature-implementation-plan.md`

---

**Document Version:** 1.0
**Last Updated:** November 22, 2025
**Status:** Ready for Implementation

🤖 Generated with [Claude Code](https://claude.com/claude-code)
