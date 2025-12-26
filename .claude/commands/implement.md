---
model: sonnet
description: Implement code from documentation specs
argument-hint: <docs/feature-name.md>
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, TodoWrite
---

# Implement - Documentation-Driven Code Implementation

Read a documentation specification and implement the code exactly as described.

**Usage**: `/implement docs/feature-name-backend.md` or `/implement docs/feature-name-frontend.md`

## Workflow

1. **Read the documentation file** specified in $ARGUMENTS
2. **Create a todo list** based on the Implementation Checklist in the doc (if present)
3. **Implement each item** following the exact specifications:
   - Use the code snippets provided in the documentation
   - Follow the file structure specified
   - Create directories and files as documented
4. **Mark todos as completed** as you finish each item
5. **Verify** the implementation compiles/builds correctly

## Implementation Rules

- **Follow the spec exactly** - don't add features not in the documentation
- **Use code from the doc** - copy code snippets as-is, they've been reviewed
- **Create files in order** - follow the Implementation Checklist sequence
- **No over-engineering** - implement only what's specified
- **Check existing patterns** - if unsure, look at similar existing code in the codebase

## For Backend (Go) Implementations

- Run `go build ./...` after creating files to verify compilation
- Run `goimports -w -local github.com/francowini/rafiki .` to fix imports
- Follow business types pattern from `business/types/`
- Follow domain pattern from existing `business/domain/*bus/` packages

## For Frontend (TypeScript) Implementations

- Run `npm run check` after creating files to verify types
- Follow component patterns from `frontend/components/features/`
- Use shadcn/ui components as specified

## Output

After implementation:
1. List all files created/modified
2. Show any compilation/type errors encountered
3. Suggest next steps (e.g., "run migrations", "restart service", "test manually")

---

Now read $ARGUMENTS and implement the specification.
