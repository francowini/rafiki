# Life Visions - Frontend Implementation

## Overview

Life Visions feature for the Rafiki dashboard. Displays values as a table with nested life visions, plus a dedicated page for managing visions.

## TypeScript Types

**File**: `frontend/lib/types.ts` (add to existing file)

```typescript
// ============================================================================
// Life Vision Types
// ============================================================================

export interface LifeVision {
  id: string;
  valueId: string;
  content: string;
  dateCreated: string;
  dateUpdated: string;
}

export interface NewLifeVision {
  valueId: string;
  content: string;
}

export interface UpdateLifeVision {
  content?: string;
  valueId?: string;
}

export interface LifeVisionListResponse {
  items: LifeVision[];
  total: number;
}
```

## API Client

**File**: `frontend/lib/api.ts` (add to existing api object)

```typescript
lifeVisions: {
  getAll: async (): Promise<LifeVisionListResponse> => {
    return fetchAPI<LifeVisionListResponse>('/v1/lifevisions');
  },

  getByValue: async (valueId: string): Promise<LifeVisionListResponse> => {
    return fetchAPI<LifeVisionListResponse>(`/v1/values/${valueId}/lifevisions`);
  },

  getById: async (id: string): Promise<LifeVision> => {
    return fetchAPI<LifeVision>(`/v1/lifevisions/${id}`);
  },

  create: async (data: NewLifeVision): Promise<LifeVision> => {
    return fetchAPI<LifeVision>('/v1/lifevisions', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  update: async (id: string, data: UpdateLifeVision): Promise<LifeVision> => {
    return fetchAPI<LifeVision>(`/v1/lifevisions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  delete: async (id: string): Promise<void> => {
    return fetchAPI<void>(`/v1/lifevisions/${id}`, {
      method: 'DELETE',
    });
  },
},
```

## Components

### 1. LifeVisionForm

**File**: `frontend/components/features/life-visions/LifeVisionForm.tsx`

Form for creating/editing life visions. Uses Sheet component (right slide-out).

**Props**:
```typescript
interface LifeVisionFormProps {
  lifeVision?: LifeVision | null;  // Edit mode if provided
  preselectedValueId?: string;      // Pre-select value in dropdown
  values: Value[];                  // Available values
  onSuccess?: () => void;
  onCancel?: () => void;
}
```

**Validation** (Zod schema):
```typescript
const lifeVisionSchema = z.object({
  content: z
    .string()
    .min(10, 'Vision must be at least 10 characters')
    .max(500, 'Vision must be less than 500 characters'),
  valueId: z.string().uuid('Please select a value'),
});
```

**UI Elements**:
- Privacy alert (rose-50 background)
- Value selector (Select component)
- Content textarea (min-h-32)
- Character guidance text
- Cancel/Save buttons

**Placeholder text**:
```
"Example: In 5 years, I see myself leading a team that builds products
that help millions of people live healthier lives..."
```

### 2. LifeVisionCard

**File**: `frontend/components/features/life-visions/LifeVisionCard.tsx`

Simple card displaying a single life vision.

**Props**:
```typescript
interface LifeVisionCardProps {
  lifeVision: LifeVision;
  onEdit: (lifeVision: LifeVision) => void;
  onDelete: (lifeVision: LifeVision) => void;
}
```

**UI Elements**:
- Content text (whitespace-pre-wrap)
- Timestamp (formatDistanceToNow)
- Edit button (Pencil icon)
- Delete button (Trash2 icon, destructive color)

### 3. LifeVisionsByValue

**File**: `frontend/components/features/life-visions/LifeVisionsByValue.tsx`

Groups life visions under a value header.

**Props**:
```typescript
interface LifeVisionsByValueProps {
  value: Value;
  visions: LifeVision[];
  onEdit: (vision: LifeVision) => void;
  onDelete: (vision: LifeVision) => void;
}
```

**UI Elements**:
- Value header: Priority badge + content + facet badge
- List of LifeVisionCards (ml-12 indent)
- Empty state with Sparkles icon

### 4. ValuesTableWithVisions (Dashboard)

**File**: `frontend/components/dashboard/ValuesTableWithVisions.tsx`

Table displaying values with nested life visions for dashboard.

**Table columns**:
| # | Value | Life Visions |
|---|-------|--------------|

**Features**:
- Fetches both values and visions on mount
- Groups visions by valueId
- Shows visions in rose-50 cards
- Empty state: "No visions yet" with Sparkles icon
- Link to "/life-visions" page

## Pages

### Life Visions Page

**File**: `frontend/app/(dashboard)/life-visions/page.tsx`

Dedicated page for managing life visions.

**State**:
```typescript
const [isFormOpen, setIsFormOpen] = useState(false);
const [isEditFormOpen, setIsEditFormOpen] = useState(false);
const [visionToEdit, setVisionToEdit] = useState<LifeVision | null>(null);
const [visionToDelete, setVisionToDelete] = useState<LifeVision | null>(null);
const [values, setValues] = useState<Value[]>([]);
const [visions, setVisions] = useState<LifeVision[]>([]);
const [isLoading, setIsLoading] = useState(true);
```

**Layout**:
```
┌─────────────────────────────────────────────────────────┐
│ ✨ Life Visions                      [+ New vision]     │
│ Define how you want to live each of your values        │
├─────────────────────────────────────────────────────────┤
│ ℹ️  Each value can have multiple visions, but we       │
│    recommend max 2 per value for clarity.              │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ [#1] Living with integrity...        [🌱 Personal]     │
│     ┌──────────────────────────────────────────┐       │
│     │ I have a daily meditation practice...    │       │
│     │                           [Edit] [Delete] │       │
│     └──────────────────────────────────────────┘       │
│                                                         │
│ [#2] Nurturing relationships...      [👨‍👩‍👧 Family]      │
│     ┌──────────────────────────────────────────┐       │
│     │ No visions yet. Click "New vision"...    │       │
│     └──────────────────────────────────────────┘       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Components used**:
- Sheet (create/edit forms)
- AlertDialog (delete confirmation)
- Alert (info banner)
- Badge (priority, facet)
- Button (actions)

## shadcn Components Required

**Already installed**:
- ✅ Sheet
- ✅ AlertDialog
- ✅ Alert
- ✅ Badge
- ✅ Button
- ✅ Select
- ✅ Textarea
- ✅ Label
- ✅ Card

**Need to install**:
```bash
cd frontend
npx shadcn@latest add table
```

## File Structure

```
frontend/
├── lib/
│   ├── types.ts                    # Add LifeVision types
│   └── api.ts                      # Add lifeVisions API
├── components/
│   ├── features/
│   │   └── life-visions/
│   │       ├── LifeVisionForm.tsx
│   │       ├── LifeVisionCard.tsx
│   │       └── LifeVisionsByValue.tsx
│   └── dashboard/
│       └── ValuesTableWithVisions.tsx
└── app/
    └── (dashboard)/
        ├── life-visions/
        │   └── page.tsx
        └── page.tsx                # Update to use ValuesTableWithVisions
```

## Navigation

Add to sidebar navigation:

```typescript
{
  title: 'Life Visions',
  href: '/life-visions',
  icon: Sparkles,
}
```

## UI Copy & Messaging

**Page title**: "Life Visions"
**Description**: "Define how you want to live each of your values"

**Info alert**:
> Each value can have multiple visions, but we recommend max 2 per value for clarity. Visions help you translate abstract values into concrete life goals.

**Form label**: "Your vision"
**Form help text**: "Describe how you want to live this value. Be specific and paint a picture (10-500 characters). We recommend max 2 visions per value for clarity."

**Empty state (per value)**: "No visions yet. Click 'New vision' to add one."
**Empty state (no values)**: "You need to create values first before adding life visions."

**Toast messages**:
- Create: "Vision created - Your life vision has been successfully created."
- Update: "Vision updated - Your life vision has been successfully updated."
- Delete: "Vision deleted - Your life vision has been successfully deleted."

## Styling

**Colors**:
- Primary: `bg-rose-600 hover:bg-rose-700`
- Heading: `text-rose-900`
- Info alert: `bg-rose-50 border-rose-200`
- Vision cards: `bg-rose-50 border-rose-200`
- Empty state text: `text-muted-foreground`

**Icons**:
- Feature icon: `Sparkles` (from lucide-react)
- Edit: `Pencil`
- Delete: `Trash2`
- Info: `Info`
- Loading: `Loader2`

## Implementation Checklist

- [ ] Add LifeVision types to `lib/types.ts`
- [ ] Add lifeVisions API to `lib/api.ts`
- [ ] Install shadcn table component
- [ ] Create `LifeVisionForm.tsx`
- [ ] Create `LifeVisionCard.tsx`
- [ ] Create `LifeVisionsByValue.tsx`
- [ ] Create `ValuesTableWithVisions.tsx`
- [ ] Create `/life-visions/page.tsx`
- [ ] Update dashboard to use ValuesTableWithVisions
- [ ] Add navigation link to sidebar
- [ ] Test CRUD operations
- [ ] Test responsive layout
