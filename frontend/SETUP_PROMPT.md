# Frontend Setup Prompt - Rafiki Thinks Management

I need you to generate a **complete and functional base** for a Next.js project deployed on Vercel, initially focused on **managing thoughts/notes** (called "thinks" in the system).

## Tech Stack
- **Framework**: Next.js 14+ (App Router)
- **UI**: shadcn/ui
- **Styling**: Tailwind CSS
- **Language**: TypeScript
- **Deployment**: Vercel

## Initial Functionality (MVP)

The project should allow:
1. ✅ **Create new thinks** (form with category selector and content textarea)
2. ✅ **View list of all thinks** created (with pagination support)

## Backend API (Already Working)

The Go backend exposes these endpoints:

### GET /v1/thinks
Get all thinks with pagination
- **Query params**: `page`, `rows`, `orderBy`
- **Response**:
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

### GET /v1/thinks/{think_id}
Get a single think by ID
- **Response**: Same as single think object above

### POST /v1/thinks
Create a new think
- **Request body**:
```json
{
  "category": "personal",
  "content": "Think content here"
}
```
- **Valid categories**: `"personal"`, `"work"`, `"ideas"`, `"learning"`, `"reflection"`
- **Response**: Same as single think object above

**API Base URL**: Configurable via environment variable

## Project Requirements

### 1. Folder Structure

Give me a **simple but production-ready** structure:
```
frontend/
├── app/                    # Next.js App Router
│   ├── layout.tsx         # Main layout
│   ├── page.tsx           # Home/Dashboard
│   └── thinks/            # Thinks section
│       ├── page.tsx       # List all thinks
│       └── [id]/          # Individual think (optional for MVP)
├── components/
│   ├── ui/                # shadcn/ui components
│   ├── features/          # Feature-specific components
│   │   ├── ThinkForm.tsx
│   │   ├── ThinkCard.tsx
│   │   └── ThinkList.tsx
│   └── layout/            # Layout components (navbar, etc)
│       ├── Header.tsx
│       └── Sidebar.tsx
├── lib/
│   ├── api.ts            # API client (fetch functions)
│   └── types.ts          # TypeScript types
├── public/               # Static assets
└── config files          # tsconfig, tailwind, etc
```

Briefly explain the purpose of each folder.

### 2. Step-by-Step Implementation Plan

Give me the **exact commands** in order:

#### Step 1: Create Project
```bash
# Exact command to create Next.js project
```

#### Step 2: Setup shadcn/ui
```bash
# Initialization commands and necessary components
```

#### Step 3: Base Structure
- Create necessary folders
- Main layout with simple navigation
- Thinks page

#### Step 4: Think Components
- Form to create new think
- List of thinks
- Individual think card

### 3. Code Examples

Provide complete code for:

#### A. Main Layout (`app/layout.tsx`)
- With simple navbar
- Dark mode toggle (optional)
- Basic navigation

#### B. API Client (`lib/api.ts`)
```typescript
// Must include:
// - Function for GET /v1/thinks (with pagination params)
// - Function for GET /v1/thinks/{id}
// - Function for POST /v1/thinks
// - Error handling
// - TypeScript return types
```

#### C. TypeScript Types (`lib/types.ts`)
```typescript
export interface Think {
  id: string;
  category: "personal" | "work" | "ideas" | "learning" | "reflection";
  content: string;
  dateCreated: string;
  dateUpdated: string;
}

export interface NewThink {
  category: string;
  content: string;
}

export interface ThinkListResponse {
  items: Think[];
  total: number;
  page: number;
  rowsPerPage: number;
}
```

#### D. Thinks Page (`app/thinks/page.tsx`)
- List of thinks
- Button to create new think
- Loading states
- Pagination controls

#### E. Create Form (`components/features/ThinkForm.tsx`)
- Using shadcn/ui components (Button, Textarea, Select, Card, Form)
- Category dropdown with valid options
- Content textarea
- Basic validation
- Submit to API

#### F. Environment Variables (`.env.example`)
```
NEXT_PUBLIC_API_URL=http://localhost:3000
```

### 4. Required shadcn/ui Components

Exact list of components to install:
```bash
npx shadcn@latest add button
npx shadcn@latest add card
npx shadcn@latest add textarea
npx shadcn@latest add select
npx shadcn@latest add form
npx shadcn@latest add dialog
npx shadcn@latest add badge
# Add any other necessary components
```

### 5. Vercel Deployment

Step by step:
1. Connect repository to Vercel
2. Configure environment variables
3. Automatic deployment

### 6. Final Checklist

Give me a checklist to verify everything works:
- [ ] Project runs locally (`npm run dev`)
- [ ] Can create a new think
- [ ] Can view all thinks
- [ ] Styles display correctly
- [ ] Successful Vercel deployment
- [ ] Environment variables configured
- [ ] CORS works with backend

## Important Considerations

- **Mobile-first**: Responsive design
- **Simple**: No over-engineering, only what's necessary
- **Production-ready**: Clean and organized code
- **Dark mode**: Include basic support
- **Error handling**: API error management
- **Loading states**: Loading indicators
- **Form validation**: Client-side validation before API calls

## Expected Output

1. **Complete folder structure** with explanation
2. **All commands** to execute in order
3. **Complete code** for all files mentioned above
4. **Deployment instructions** for Vercel
5. **Screenshots or description** of how the UI should look

---

**Final Goal**: Have a Next.js project running on Vercel where I can create and view thinks, connected to my Go backend, ready in less than 1 hour.
