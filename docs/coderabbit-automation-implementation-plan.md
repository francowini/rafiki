# CodeRabbit Automation Implementation Plan

**Project**: Rafiki Habits Tracker
**Date**: 2025-11-17
**Status**: Planning Phase
**Specialists**: Backend Engineer + Frontend Engineer + DevOps Engineer

---

## Executive Summary

This document provides a comprehensive implementation plan for integrating CodeRabbit code review automation into the Rafiki development workflow. The solution combines:

1. **Automated Tier 1 fixes** via pre-commit hooks and GitHub Actions (Blacksmith CI)
2. **Interactive Tier 2/3 review** via Claude Code slash command
3. **GitHub branch protection** to enforce code quality standards

The system will automatically handle ~80% of formatting/linting issues while providing an interactive workflow for logic, security, and architectural changes that require human judgment.

---

## Table of Contents

1. [Cross-Pollination Analysis](#cross-pollination-analysis)
2. [Architecture Overview](#architecture-overview)
3. [Three-Tier Fix Strategy](#three-tier-fix-strategy)
4. [Implementation Components](#implementation-components)
5. [Claude Code Slash Command Design](#claude-code-slash-command-design)
6. [Pre-Commit Hooks Strategy](#pre-commit-hooks-strategy)
7. [GitHub Actions with Blacksmith CI](#github-actions-with-blacksmith-ci)
8. [GitHub Branch Protection](#github-branch-protection)
9. [Configuration Files](#configuration-files)
10. [Workflow Examples](#workflow-examples)
11. [Dependencies and Coordination](#dependencies-and-coordination)
12. [Deployment Strategy](#deployment-strategy)
13. [Risk Analysis and Mitigation](#risk-analysis-and-mitigation)
14. [Success Metrics](#success-metrics)
15. [Implementation Roadmap](#implementation-roadmap)

---

## Cross-Pollination Analysis

### Backend ↔ Frontend Integration Points

#### Shared Concerns
Both backend (Go) and frontend (Next.js/TypeScript) share similar auto-fix needs:

| Concern | Backend Solution | Frontend Solution | Integration Point |
|---------|------------------|-------------------|-------------------|
| **Formatting** | gofmt, gofumpt | Prettier | Run both in parallel in CI |
| **Linting** | golangci-lint | ESLint | Separate jobs, unified reporting |
| **Import Organization** | goimports | ESLint + Prettier | Consistent ordering philosophy |
| **Type Safety** | Strong business types | TypeScript strict mode | API contract validation |
| **Error Handling** | Custom error wrapping | Type-safe API errors | Shared error code definitions |

#### API Contract Synchronization

**Critical Insight**: Changes to backend API contracts require frontend type updates.

**Backend Impact on Frontend:**
- API endpoint signature changes → Frontend API client types must update
- Error response format changes → Frontend error handling must adapt
- New fields in domain models → Frontend TypeScript interfaces need updates

**Frontend Impact on Backend:**
- New UI requirements → May need new API endpoints
- Form validation rules → Backend must enforce same rules
- Authentication flow changes → Backend auth logic must support

**Solution**: The Claude Code slash command should detect API contract changes and:
1. Flag them as Tier 3 (manual review)
2. Present both backend and frontend files that need coordination
3. Suggest creating a coordinated commit that updates both

### DevOps ↔ Development Integration

#### Pre-Commit Hooks (Local Development)

**Backend Hooks:**
```bash
# .git/hooks/pre-commit (Go portion)
gofmt -w -s .
goimports -w -local github.com/francowini/rafiki .
golangci-lint run --fix --enable-only=gofmt,goimports,gci,gosimple
```

**Frontend Hooks:**
```bash
# .git/hooks/pre-commit (Frontend portion)
cd frontend
npm run format
npm run lint:fix
npm run typecheck
```

**Integration Challenge**: Single pre-commit hook must handle both
**Solution**: Smart detection of changed files and run only relevant tools

#### GitHub Actions with Blacksmith CI

**Blacksmith Benefits:**
- 2-5x faster than standard GitHub Actions runners
- Better for compute-intensive tasks (linting, testing)
- Cost-effective for private repositories

**Parallel Job Strategy:**
```yaml
jobs:
  autofix-backend:
    runs-on: blacksmith-2vcpu-ubuntu-2204  # Fast runner
    # Go auto-fixes

  autofix-frontend:
    runs-on: blacksmith-2vcpu-ubuntu-2204  # Fast runner
    # Frontend auto-fixes

  # Jobs run in parallel, results aggregated
```

### Cross-Codebase Concerns

#### Security-Sensitive Files (Tier 3 - Never Auto-Fix)

**Backend:**
- `business/domain/userbus/` - User management, password hashing
- `app/sdk/auth/` - JWT token handling
- `foundation/keystore/` - Cryptographic keys
- `business/types/` - Business validation logic

**Frontend:**
- `lib/auth-context.tsx` - Authentication state
- `lib/api.ts` - API client with token handling
- `app/api/` - API routes (if any)
- `components/auth/` - Login/signup forms

**Coordination Requirement**: Changes to auth flow require both backend and frontend updates in sync.

#### Performance Optimization (Tier 3)

**Backend:**
- Database query optimization
- Concurrency patterns
- HTTP middleware ordering

**Frontend:**
- React component memoization
- Image optimization
- Bundle splitting
- API call batching

**Why Tier 3**: Performance changes need profiling data and load testing, not suitable for automation.

---

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Developer Workflow                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │  1. Create PR on GitHub                 │
         └────────────────────────────────────────┘
                              │
         ┌────────────────────┴────────────────────┐
         ▼                                         ▼
┌─────────────────────┐                  ┌─────────────────────┐
│  CodeRabbit Pro     │                  │  GitHub Actions     │
│  (GitHub App)       │                  │  (Blacksmith CI)    │
│                     │                  │                     │
│  • Auto-reviews PR  │                  │  • Tier 1 auto-fix  │
│  • Posts comments   │                  │  • Run tests        │
│  • Suggestions      │                  │  • Run linters      │
└─────────────────────┘                  └─────────────────────┘
         │                                         │
         │                                         │
         └────────────────────┬────────────────────┘
                              ▼
         ┌────────────────────────────────────────┐
         │  Branch Protection (GitHub)             │
         │  • Requires status checks               ���
         │  • Requires reviews                     │
         │  • Blocks direct push to main           │
         └────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │  Developer reviews CodeRabbit comments  │
         └────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │  Claude Code Slash Command              │
         │  `/coderabbit-review`                   │
         │                                         │
         │  1. Fetch CodeRabbit comments (API)     │
         │  2. Categorize by Tier (1/2/3)          │
         │  3. Auto-apply Tier 1 fixes             │
         │  4. Present Tier 2 interactively        │
         │  5. Flag Tier 3 for manual review       │
         │  6. Commit and push fixes               │
         └────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │  PR updated → CodeRabbit re-reviews     │
         └────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │  Approval & Merge                       │
         └────────────────────────────────────────┘
```

### Three Enforcement Layers

1. **Local (Pre-Commit Hooks)** - Fastest feedback, runs before code leaves developer's machine
2. **CI/CD (GitHub Actions + Blacksmith)** - Enforced for all PRs, catches what hooks missed
3. **Branch Protection** - Final gatekeeper, ensures nothing bypasses quality gates

---

## Three-Tier Fix Strategy

### Tier 1: Safe Auto-Fix (No Human Approval)

**Criteria**: Pure formatting/style changes with zero logic impact

#### Backend (Go)

**Tools**: `gofmt`, `goimports`, `gofumpt`, `golangci-lint` (limited linters)

**Auto-Fixable Issues**:
- Code formatting (indentation, spacing, line breaks)
- Import organization and grouping
- Unused imports removal
- Semicolon/brace placement (Go convention)
- Simple code simplifications (e.g., `if err != nil { return false } return true` → `return err == nil`)
- Whitespace cleanup
- Canonical header formatting

**Excluded from Tier 1**:
- Any changes to `business/types/*` (business validation logic)
- Any changes to `business/domain/*/model.go` (domain models)
- Error handling logic
- Exported function signatures
- Database queries

**Command**:
```bash
# Backend Tier 1 auto-fix
gofmt -w -s $(find . -name "*.go" -not -path "./vendor/*")
goimports -w -local github.com/francowini/rafiki .
golangci-lint run --fix --enable-only=gofmt,goimports,gci,gosimple,ineffassign,whitespace
```

#### Frontend (Next.js/TypeScript)

**Tools**: `prettier`, `eslint --fix` (safe rules only)

**Auto-Fixable Issues**:
- Code formatting (indentation, line length, wrapping)
- Quote consistency (single vs double)
- Semicolon usage
- Trailing commas
- Import order and organization
- Unused imports removal
- Spacing and bracket formatting
- JSX prop formatting

**Excluded from Tier 1**:
- Any changes to `lib/auth-context.tsx` (auth logic)
- Any changes to `lib/api.ts` (API client)
- Component logic changes
- Type definitions (replacing `any`)
- Event handlers
- State management

**Command**:
```bash
# Frontend Tier 1 auto-fix
cd frontend
prettier --write "**/*.{ts,tsx,json,md}"
eslint . --ext .ts,.tsx --fix --rule 'no-console: off' --rule '@typescript-eslint/no-explicit-any: off'
```

**Configuration Required**:
- Create `frontend/.prettierrc`
- Update `frontend/eslint.config.mjs` with auto-fix safe rules
- Create `frontend/.prettierignore`

### Tier 2: Requires Approval (Interactive Review)

**Criteria**: Logic changes that are likely correct but need human verification

#### Backend Examples:

**Error Handling Improvements**:
```go
// CodeRabbit suggests
- fmt.Errorf("error: %s", err)
+ fmt.Errorf("error: %w", err)
```
**Reason for approval**: Changes error wrapping behavior, affects error inspection

**Unused Code Removal**:
```go
// CodeRabbit suggests removing
func helperFunction() { ... }  // Not called anywhere
```
**Reason for approval**: Might be intended for future use, or used via reflection

**Code Simplification**:
```go
// CodeRabbit suggests
- if len(items) == 0 { return nil }
+ if len(items) == 0 { return nil }  // No change, but...
- return processItems(items)
+ return processItems(items)
```
**Reason for approval**: Simplifications might change semantics subtly

#### Frontend Examples:

**Type Safety Improvements**:
```typescript
// CodeRabbit suggests
- } catch (err: any) {
+ } catch (err: unknown) {
+   const error = err as Error;
```
**Reason for approval**: Changes error handling pattern, needs verification

**Missing Dependencies in Hooks**:
```typescript
// CodeRabbit suggests
useEffect(() => {
  fetchData(userId);
- }, []);
+ }, [userId]);
```
**Reason for approval**: Could cause infinite re-renders if userId changes frequently

**Console Statement Removal**:
```typescript
// CodeRabbit suggests removing
- console.error('Login error:', err);
```
**Reason for approval**: Might want to replace with proper logger, not just delete

#### Interactive Workflow:

For each Tier 2 issue, the Claude Code command should:
1. **Present the suggestion** with file, line number, and context
2. **Show the diff** (before/after)
3. **Explain CodeRabbit's reasoning**
4. **Ask the developer**: "Apply this fix? (yes/no/skip-all-similar)"
5. **Record decision** and apply if approved

### Tier 3: Manual Only (Never Auto-Fix)

**Criteria**: Security-sensitive, business logic, or architectural changes

#### Backend - Never Auto-Fix:

**Business Logic** (`business/types/*`, `business/domain/*`):
```go
// Example: business/types/intensity/intensity.go
func Parse(value int) (Intensity, error) {
    if value < 0 || value > 10 {  // NEVER auto-change validation rules
        return Intensity{}, fmt.Errorf("...")
    }
}
```

**Security Code** (`app/sdk/auth/*`, `foundation/keystore/*`):
```go
// Example: JWT token generation, password hashing
// ANY change here requires manual security review
```

**API Contracts** (`app/domain/*/model.go`):
```go
// Example: API request/response models
// Changes affect frontend integration
```

**Database Operations** (`business/domain/*/stores/*`):
```go
// Example: SQL queries, migrations
// Changes affect data integrity
```

**Concurrency Patterns**:
```go
// Example: goroutines, channels, mutexes
// Changes could introduce race conditions
```

#### Frontend - Never Auto-Fix:

**Authentication Logic** (`lib/auth-context.tsx`, `components/auth/*`):
```typescript
// Example: Token storage, session management
// Security implications require manual review
```

**API Client** (`lib/api.ts`):
```typescript
// Example: Request/response handling, error mapping
// Changes affect backend integration
```

**Component Business Logic**:
```typescript
// Example: Form validation, data transformations
// UX implications require manual review
```

**Performance Optimizations**:
```typescript
// Example: useMemo, useCallback, React.memo
// Could break functionality if applied incorrectly
```

#### Tier 3 Workflow:

The Claude Code command should:
1. **Flag these issues** with a ⚠️ warning
2. **Explain why** it's Tier 3 (security/business logic/etc.)
3. **Provide context** but NOT suggest auto-fix
4. **Create a checklist** for manual review
5. **Skip** to next issue

---

## Implementation Components

### Component 1: CodeRabbit Pro Setup

**Status**: ✅ Already installed (user confirmed)

**Configuration Needed**: Create `.coderabbit.yaml` at repository root

**Key Features to Configure**:
- Review profile (assertive mode)
- Path filters (exclude `vendor/`, `node_modules/`, `.next/`)
- Tool integrations (golangci-lint, ESLint)
- Path-specific instructions (business logic patterns)
- Custom review rules for Rafiki patterns

### Component 2: GitHub Branch Protection

**Status**: ⏳ To be configured

**Settings**:
- Branch: `main`
- Require pull request reviews: 1 approval
- Require status checks: ✅
  - `lint-backend`
  - `lint-frontend`
  - `test-backend`
  - `test-frontend`
  - `typecheck-frontend`
- Require conversation resolution: ✅
- Require linear history: ✅
- Block direct pushes: ✅
- Enforce for administrators: ✅

**Configuration Method**: GitHub UI or API script

### Component 3: Pre-Commit Hooks

**Status**: ⏳ To be created

**Tool**: Husky + lint-staged (for frontend) + custom script (for backend)

**Strategy**:
- Detect changed files
- Run Go tools on `.go` files
- Run frontend tools on `.ts`/`.tsx` files
- Fail commit if errors found
- Allow `--no-verify` bypass for emergencies

### Component 4: GitHub Actions Workflows

**Status**: ⏳ To be created

**Workflows Needed**:

1. **`pr-autofix.yml`** - Auto-fix Tier 1 issues on PR push
2. **`pr-checks.yml`** - Run tests and linting (required status checks)
3. **`blacksmith-lint.yml`** - Fast linting with Blacksmith CI

**Blacksmith Integration**:
```yaml
runs-on: blacksmith-2vcpu-ubuntu-2204  # Instead of ubuntu-latest
```

### Component 5: Claude Code Slash Command

**Status**: ⏳ To be created

**Location**: `.claude/commands/coderabbit-review.md`

**Functionality**:
1. Detect current branch and PR number
2. Fetch CodeRabbit comments via GitHub API
3. Categorize each comment by Tier
4. Auto-apply Tier 1 fixes
5. Present Tier 2 interactively (ask user for each)
6. Flag Tier 3 for manual review
7. Commit and push changes
8. Post summary comment on PR

**Command Invocation**:
```bash
# In Claude Code CLI
/coderabbit-review

# With options
/coderabbit-review --pr=123 --auto-tier1 --interactive
```

### Component 6: Configuration Files

**Status**: ⏳ To be created

**Files Needed**:
1. `.coderabbit.yaml` - CodeRabbit configuration
2. `.golangci.yml` - Go linter configuration
3. `frontend/.prettierrc` - Prettier configuration
4. `frontend/.prettierignore` - Prettier ignore patterns
5. `frontend/eslint.config.mjs` - Enhanced ESLint config
6. `.github/workflows/pr-autofix.yml` - Auto-fix workflow
7. `.github/workflows/pr-checks.yml` - Status checks
8. `.github/workflows/blacksmith-lint.yml` - Blacksmith linting
9. `.claude/commands/coderabbit-review.md` - Slash command
10. `scripts/setup-branch-protection.sh` - Branch protection script

---

## Claude Code Slash Command Design

### Command File Structure

**File**: `.claude/commands/coderabbit-review.md`

### Command Metadata

```markdown
# CodeRabbit Review

Fetch CodeRabbit PR comments and interactively address them with tiered auto-fix strategy.

**Usage**: `/coderabbit-review [--pr=NUMBER] [--tier1-only]`

## Arguments

- `--pr=NUMBER` (optional): Specify PR number. If omitted, detects from current branch.
- `--tier1-only` (optional): Only apply Tier 1 safe auto-fixes, skip interactive review.
- `--dry-run` (optional): Show what would be fixed without applying changes.

## How It Works

1. **Detect PR**: Finds PR associated with current branch
2. **Fetch Comments**: Retrieves CodeRabbit review comments via GitHub API
3. **Categorize**: Sorts comments into Tier 1 (safe), Tier 2 (review), Tier 3 (manual)
4. **Tier 1 Auto-Fix**: Applies formatting and linting fixes automatically
5. **Tier 2 Interactive**: Presents each suggestion and asks for approval
6. **Tier 3 Checklist**: Creates a manual review checklist
7. **Commit & Push**: Commits fixes and pushes to PR branch
8. **Summary**: Posts a comment on the PR with what was fixed

## Example

\`\`\`bash
# Review current PR
/coderabbit-review

# Review specific PR
/coderabbit-review --pr=42

# Only apply safe auto-fixes
/coderabbit-review --tier1-only

# Preview changes without applying
/coderabbit-review --dry-run
\`\`\`
```

### Command Implementation Logic

**Phase 1: Context Detection**

```bash
# Detect current branch
CURRENT_BRANCH=$(git branch --show-current)

# Detect PR number from branch
PR_NUMBER=$(gh pr list --head "$CURRENT_BRANCH" --json number --jq '.[0].number')

# If PR doesn't exist, error
if [ -z "$PR_NUMBER" ]; then
  echo "❌ No PR found for branch $CURRENT_BRANCH"
  exit 1
fi

echo "📋 Found PR #$PR_NUMBER"
```

**Phase 2: Fetch CodeRabbit Comments**

```bash
# Fetch all review comments from PR
COMMENTS=$(gh api repos/{owner}/{repo}/pulls/$PR_NUMBER/comments \
  --jq '.[] | select(.user.login == "coderabbitai[bot]")')

# Extract comments with suggestions
SUGGESTIONS=$(echo "$COMMENTS" | jq -r 'select(.body | contains("```suggestion"))')

# Count total suggestions
TOTAL=$(echo "$SUGGESTIONS" | jq -s 'length')

echo "🤖 Found $TOTAL CodeRabbit suggestions"
```

**Phase 3: Categorization Logic**

The command must analyze each comment and categorize based on:

**Tier 1 Indicators** (safe auto-fix):
- File path NOT in protected directories
- Comment body contains formatting keywords: "formatting", "indentation", "import", "whitespace"
- Suggestion is pure cosmetic (no logic change)
- Changes to non-exported code only

**Tier 2 Indicators** (requires approval):
- File path in business logic directories
- Comment mentions: "error handling", "simplify", "remove unused"
- Changes to exported functions/types
- Logic changes with clear benefit

**Tier 3 Indicators** (manual only):
- File path in security-sensitive directories:
  - `business/types/*`
  - `business/domain/*/model.go`
  - `app/sdk/auth/*`
  - `lib/auth-context.tsx`
  - `lib/api.ts`
- Comment mentions: "security", "performance", "concurrency", "database"
- API contract changes
- Complex refactoring

**Categorization Implementation**:

```javascript
// Pseudo-code for categorization logic
function categorizeSuggestion(comment) {
  const filePath = comment.path;
  const body = comment.body.toLowerCase();
  const diffLines = extractDiff(comment.body);

  // Tier 3 (highest priority - manual only)
  if (
    filePath.includes('business/types/') ||
    filePath.includes('business/domain/') && filePath.endsWith('model.go') ||
    filePath.includes('app/sdk/auth/') ||
    filePath.includes('foundation/keystore/') ||
    filePath.includes('lib/auth-context.tsx') ||
    filePath.includes('lib/api.ts') ||
    body.includes('security') ||
    body.includes('sql') ||
    body.includes('jwt') ||
    body.includes('password')
  ) {
    return { tier: 3, reason: 'Security-sensitive or business logic file' };
  }

  // Tier 2 (requires approval)
  if (
    body.includes('error handling') ||
    body.includes('unused') && !body.includes('import') ||
    body.includes('simplify logic') ||
    body.includes('refactor') ||
    diffLines.some(line => line.includes('func ') && line.includes('(')) // Function signature change
  ) {
    return { tier: 2, reason: 'Logic change requires review' };
  }

  // Tier 1 (safe auto-fix) - default for everything else
  if (
    body.includes('formatting') ||
    body.includes('import') ||
    body.includes('whitespace') ||
    body.includes('indentation') ||
    body.includes('prettier') ||
    body.includes('eslint')
  ) {
    return { tier: 1, reason: 'Safe formatting/style fix' };
  }

  // Conservative default: if unsure, require approval
  return { tier: 2, reason: 'Uncertain categorization, requires review' };
}
```

**Phase 4: Tier 1 Auto-Fix**

```bash
echo "🔧 Applying Tier 1 safe auto-fixes..."

# Backend auto-fix
if [ -n "$(echo "$TIER1_COMMENTS" | jq -r 'select(.path | endswith(".go"))')" ]; then
  echo "  → Running Go formatters..."
  gofmt -w -s .
  goimports -w -local github.com/francowini/rafiki .
  golangci-lint run --fix --enable-only=gofmt,goimports,gci,gosimple,ineffassign,whitespace || true
fi

# Frontend auto-fix
if [ -n "$(echo "$TIER1_COMMENTS" | jq -r 'select(.path | startswith("frontend/"))')" ]; then
  echo "  → Running frontend formatters..."
  cd frontend
  npm run format || true
  npm run lint:fix || true
  cd ..
fi

# Check if any changes were made
if [ -n "$(git status --porcelain)" ]; then
  echo "✅ Tier 1 fixes applied"
  TIER1_APPLIED=true
else
  echo "ℹ️  No Tier 1 changes needed"
  TIER1_APPLIED=false
fi
```

**Phase 5: Tier 2 Interactive Review**

```javascript
// For each Tier 2 comment, present interactively to user
for (const comment of tier2Comments) {
  console.log(`\n${'='.repeat(80)}`);
  console.log(`📍 ${comment.path}:${comment.line}`);
  console.log(`${'='.repeat(80)}\n`);

  // Extract suggestion from comment body
  const suggestionMatch = comment.body.match(/```suggestion\n([\s\S]*?)\n```/);
  const originalMatch = comment.body.match(/```[a-z]*\n([\s\S]*?)\n```suggestion/);

  if (suggestionMatch && originalMatch) {
    console.log('📝 CodeRabbit suggests:\n');
    console.log('BEFORE:');
    console.log(originalMatch[1]);
    console.log('\nAFTER:');
    console.log(suggestionMatch[1]);
    console.log('\n💬 Reason:', comment.body.split('\n')[0]); // First line is usually the reason
  }

  // Ask user
  const response = await askUser(
    'Apply this fix? [y]es / [n]o / [s]kip remaining / [v]iew full context',
    ['y', 'n', 's', 'v']
  );

  if (response === 'y') {
    // Apply the fix
    applyCommentSuggestion(comment);
    tier2Applied.push(comment);
  } else if (response === 's') {
    // Skip all remaining Tier 2
    break;
  } else if (response === 'v') {
    // Show full file context
    showFileContext(comment.path, comment.line, 10); // 10 lines before/after
    // Re-ask
    continue;
  }
  // 'n' = skip this one, continue to next
}
```

**Phase 6: Tier 3 Manual Checklist**

```bash
if [ ${#TIER3_COMMENTS[@]} -gt 0 ]; then
  echo ""
  echo "⚠️  Manual Review Required (Tier 3)"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "The following issues require manual review due to security or business logic concerns:"
  echo ""

  for comment in "${TIER3_COMMENTS[@]}"; do
    FILE=$(echo "$comment" | jq -r '.path')
    LINE=$(echo "$comment" | jq -r '.line')
    REASON=$(echo "$comment" | jq -r '.categorization.reason')

    echo "- [ ] $FILE:$LINE"
    echo "      Reason: $REASON"
    echo "      Link: $(echo "$comment" | jq -r '.html_url')"
    echo ""
  done

  # Create a GitHub issue for tracking (optional)
  read -p "Create GitHub issue for Tier 3 review? (y/n) " -n 1 -r
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Create issue with checklist
    gh issue create --title "[CodeRabbit] Manual Review Required - PR #$PR_NUMBER" \
      --body "$(generate_tier3_checklist)"
  fi
fi
```

**Phase 7: Commit and Push**

```bash
if [ "$TIER1_APPLIED" = true ] || [ ${#TIER2_APPLIED[@]} -gt 0 ]; then
  echo ""
  echo "💾 Committing changes..."

  git add .

  # Generate commit message
  COMMIT_MSG="fix: address CodeRabbit suggestions for PR #$PR_NUMBER

Applied fixes:
- Tier 1 (auto): ${TIER1_COUNT} formatting/linting issues
- Tier 2 (reviewed): ${#TIER2_APPLIED[@]} logic improvements

Tier 3 manual review: ${#TIER3_COMMENTS[@]} issues flagged

🤖 Generated with Claude Code /coderabbit-review command

Co-Authored-By: CodeRabbit <noreply@coderabbit.ai>"

  git commit -m "$COMMIT_MSG"

  echo "📤 Pushing to branch $CURRENT_BRANCH..."
  git push origin "$CURRENT_BRANCH"

  echo "✅ Changes pushed to PR #$PR_NUMBER"
else
  echo "ℹ️  No changes to commit"
fi
```

**Phase 8: PR Summary Comment**

```bash
# Post summary comment to PR
gh pr comment "$PR_NUMBER" --body "$(cat <<EOF
## 🤖 CodeRabbit Review Summary

**Processed**: $(date)
**Command**: \`/coderabbit-review\`

### Fixes Applied

- ✅ **Tier 1 (Auto-fix)**: $TIER1_COUNT issues
  - Formatting, linting, import organization

- ✅ **Tier 2 (Reviewed)**: ${#TIER2_APPLIED[@]} issues
  $(for comment in "${TIER2_APPLIED[@]}"; do
    echo "  - $(basename $(echo "$comment" | jq -r '.path')):$(echo "$comment" | jq -r '.line')"
  done)

- ⚠️ **Tier 3 (Manual Review)**: ${#TIER3_COMMENTS[@]} issues
  - Security-sensitive or business logic changes
  - Please review manually

### Next Steps

- CodeRabbit will re-review the updated PR
- Address remaining Tier 3 issues manually
- Request review when ready

---
*Generated by Claude Code - [View workflow documentation](docs/coderabbit-automation-implementation-plan.md)*
EOF
)"
```

### Command Error Handling

```bash
# Handle common errors gracefully

# No PR found
if [ -z "$PR_NUMBER" ]; then
  echo "❌ Error: No PR found for current branch"
  echo "💡 Tip: Create a PR first with: gh pr create"
  exit 1
fi

# No CodeRabbit comments
if [ "$TOTAL_COMMENTS" -eq 0 ]; then
  echo "✅ No CodeRabbit suggestions found"
  echo "💡 Either CodeRabbit hasn't reviewed yet, or the code is perfect! 🎉"
  exit 0
fi

# Git conflicts
if ! git diff --quiet; then
  echo "⚠️  Warning: You have uncommitted changes"
  echo "💡 Tip: Commit or stash changes before running /coderabbit-review"
  read -p "Continue anyway? (y/n) " -n 1 -r
  echo
  [[ ! $REPLY =~ ^[Yy]$ ]] && exit 1
fi

# GitHub API rate limit
if [ "$(gh api rate_limit --jq '.resources.core.remaining')" -lt 10 ]; then
  echo "⚠️  Warning: GitHub API rate limit is low"
  echo "💡 Wait a few minutes or authenticate with a PAT token"
fi
```

---

## Pre-Commit Hooks Strategy

### Option 1: Husky (Node-based, Frontend-focused)

**Pros**: Well-integrated with npm ecosystem
**Cons**: Requires Node.js for Go projects (not ideal)

### Option 2: Custom Git Hook Script (Recommended)

**Location**: `.git/hooks/pre-commit` (not checked into repo)

**Installation**: Via setup script that developers run once

**Script**: `scripts/install-git-hooks.sh`

```bash
#!/bin/bash
# Install pre-commit hooks for Rafiki

HOOK_FILE=".git/hooks/pre-commit"

cat > "$HOOK_FILE" << 'EOF'
#!/bin/bash
# Rafiki Pre-Commit Hook
# Auto-format code before committing

set -e

echo "🔍 Running pre-commit checks..."

# Get list of staged files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)
STAGED_TS_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(ts|tsx)$' | grep '^frontend/' || true)

# Backend: Go files
if [ -n "$STAGED_GO_FILES" ]; then
  echo "  → Formatting Go files..."

  # Run gofmt
  echo "$STAGED_GO_FILES" | xargs gofmt -w -s

  # Run goimports
  echo "$STAGED_GO_FILES" | xargs goimports -w -local github.com/francowini/rafiki

  # Re-stage formatted files
  echo "$STAGED_GO_FILES" | xargs git add

  echo "  ✅ Go files formatted"
fi

# Frontend: TypeScript files
if [ -n "$STAGED_TS_FILES" ]; then
  echo "  → Formatting TypeScript files..."

  cd frontend

  # Run Prettier
  echo "$STAGED_TS_FILES" | sed 's|^frontend/||' | xargs npx prettier --write

  # Run ESLint fix
  echo "$STAGED_TS_FILES" | sed 's|^frontend/||' | xargs npx eslint --fix || true

  cd ..

  # Re-stage formatted files
  echo "$STAGED_TS_FILES" | xargs git add

  echo "  ✅ TypeScript files formatted"
fi

# Run quick lint check (fail fast)
if [ -n "$STAGED_GO_FILES" ]; then
  echo "  → Quick Go lint check..."
  golangci-lint run --fast --new-from-rev=HEAD~1 || {
    echo "❌ Linting failed. Fix issues or use 'git commit --no-verify' to bypass."
    exit 1
  }
fi

if [ -n "$STAGED_TS_FILES" ]; then
  echo "  → Quick TypeScript check..."
  cd frontend
  npx tsc --noEmit || {
    echo "❌ Type check failed. Fix issues or use 'git commit --no-verify' to bypass."
    exit 1
  }
  cd ..
fi

echo "✅ Pre-commit checks passed"
EOF

chmod +x "$HOOK_FILE"

echo "✅ Git hooks installed successfully"
echo "💡 To bypass hooks, use: git commit --no-verify"
```

**Developer Setup**:
```bash
# One-time setup per developer
./scripts/install-git-hooks.sh
```

### Hook Bypass for Emergencies

```bash
# Bypass hooks when needed (e.g., WIP commits)
git commit --no-verify -m "WIP: work in progress"

# Note: CI will still catch issues
```

---

## GitHub Actions with Blacksmith CI

### Why Blacksmith?

**Benefits**:
- 2-5x faster runners (more CPU cores)
- Better cache performance
- Cost-effective for private repos
- Same pricing as GitHub Actions for public repos

**Setup**:
1. Install Blacksmith GitHub App
2. Replace `runs-on: ubuntu-latest` with `runs-on: blacksmith-2vcpu-ubuntu-2204`

### Workflow 1: PR Auto-Fix (Tier 1)

**File**: `.github/workflows/pr-autofix.yml`

```yaml
name: PR Auto-Fix (Tier 1)

on:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened]

permissions:
  contents: write
  pull-requests: write

concurrency:
  group: autofix-${{ github.head_ref }}
  cancel-in-progress: true

jobs:
  autofix-backend:
    name: Auto-fix Go Code
    runs-on: blacksmith-2vcpu-ubuntu-2204
    if: contains(github.event.pull_request.changed_files, '.go')

    steps:
      - name: Checkout PR branch
        uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}
          token: ${{ secrets.GITHUB_TOKEN }}
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25.3'
          cache: true

      - name: Install Go tools
        run: |
          go install golang.org/x/tools/cmd/goimports@latest
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

      - name: Run Go auto-fix (Tier 1 only)
        run: |
          # Format
          gofmt -w -s $(find . -name "*.go" -not -path "./vendor/*")

          # Imports
          goimports -w -local github.com/francowini/rafiki $(find . -name "*.go" -not -path "./vendor/*")

          # Safe linting fixes only
          golangci-lint run --fix \
            --enable-only=gofmt,goimports,gci,gosimple,ineffassign,whitespace \
            --timeout=10m || true

      - name: Check for changes
        id: check
        run: |
          if [ -n "$(git status --porcelain)" ]; then
            echo "changed=true" >> $GITHUB_OUTPUT
          fi

      - name: Commit and push
        if: steps.check.outputs.changed == 'true'
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add .
          git commit -m "style: auto-fix Go formatting (Tier 1)

          🤖 Automated fixes by GitHub Actions:
          - gofmt formatting
          - goimports organization
          - golangci-lint safe fixes"
          git push

  autofix-frontend:
    name: Auto-fix Frontend Code
    runs-on: blacksmith-2vcpu-ubuntu-2204
    if: contains(github.event.pull_request.changed_files, 'frontend/')

    steps:
      - name: Checkout PR branch
        uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}
          token: ${{ secrets.GITHUB_TOKEN }}
          fetch-depth: 0

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run frontend auto-fix (Tier 1 only)
        working-directory: frontend
        run: |
          # Prettier
          npm run format || true

          # ESLint safe fixes
          npm run lint:fix || true

      - name: Check for changes
        id: check
        run: |
          if [ -n "$(git status --porcelain)" ]; then
            echo "changed=true" >> $GITHUB_OUTPUT
          fi

      - name: Commit and push
        if: steps.check.outputs.changed == 'true'
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add frontend/
          git commit -m "style: auto-fix frontend formatting (Tier 1)

          🤖 Automated fixes by GitHub Actions:
          - Prettier formatting
          - ESLint auto-fixes"
          git push
```

### Workflow 2: PR Status Checks (Required)

**File**: `.github/workflows/pr-checks.yml`

```yaml
name: PR Checks

on:
  pull_request:
    branches: [main]

permissions:
  contents: read
  pull-requests: write
  statuses: write

jobs:
  lint-backend:
    name: Lint Go Code
    runs-on: blacksmith-2vcpu-ubuntu-2204

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.3'
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=10m --config=.golangci.yml

  test-backend:
    name: Test Go Code
    runs-on: blacksmith-4vcpu-ubuntu-2204  # More CPU for tests

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.3'
          cache: true

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  lint-frontend:
    name: Lint Frontend
    runs-on: blacksmith-2vcpu-ubuntu-2204

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - working-directory: frontend
        run: npm ci

      - name: ESLint
        working-directory: frontend
        run: npm run lint

      - name: Prettier check
        working-directory: frontend
        run: npm run format:check

  typecheck-frontend:
    name: TypeScript Type Check
    runs-on: blacksmith-2vcpu-ubuntu-2204

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - working-directory: frontend
        run: npm ci

      - name: Type check
        working-directory: frontend
        run: npm run typecheck

  test-frontend:
    name: Build Frontend
    runs-on: blacksmith-2vcpu-ubuntu-2204

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - working-directory: frontend
        run: npm ci

      - name: Build
        working-directory: frontend
        run: npm run build
```

---

## GitHub Branch Protection

### Configuration via GitHub UI

**Navigation**: Settings → Branches → Branch protection rules → Add rule

**Branch name pattern**: `main`

**Settings**:

```yaml
Protect matching branches:
  ☑ Require a pull request before merging
    - Required approvals: 1
    - Dismiss stale pull request approvals when new commits are pushed: ☑
    - Require review from Code Owners: ☐ (optional)
    - Restrict who can dismiss pull request reviews: ☐
    - Allow specified actors to bypass required pull requests: ☐

  ☑ Require status checks to pass before merging
    - Require branches to be up to date before merging: ☑
    - Status checks that are required:
      • lint-backend
      • test-backend
      • lint-frontend
      • typecheck-frontend
      • test-frontend

  ☑ Require conversation resolution before merging

  ☑ Require signed commits: ☐ (optional but recommended)

  ☑ Require linear history

  ☑ Require deployments to succeed before merging: ☐

  ☐ Lock branch

  ☑ Do not allow bypassing the above settings

  ☑ Restrict who can push to matching branches
    - Leave empty to block all direct pushes

  ☑ Allow force pushes: ☐ NEVER

  ☑ Allow deletions: ☐ NEVER
```

### Configuration via API Script

**File**: `scripts/setup-branch-protection.sh`

```bash
#!/bin/bash
# Setup GitHub branch protection for main branch

set -e

# Check if GITHUB_TOKEN is set
if [ -z "$GITHUB_TOKEN" ]; then
  echo "❌ Error: GITHUB_TOKEN environment variable not set"
  echo "💡 Create a token at https://github.com/settings/tokens"
  echo "   Required scopes: repo, admin:repo_hook"
  exit 1
fi

OWNER="francowini"
REPO="rafiki"
BRANCH="main"

echo "🔒 Setting up branch protection for $OWNER/$REPO:$BRANCH"

# Apply branch protection
curl -X PUT \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/$OWNER/$REPO/branches/$BRANCH/protection" \
  -d '{
    "required_status_checks": {
      "strict": true,
      "contexts": [
        "lint-backend",
        "test-backend",
        "lint-frontend",
        "typecheck-frontend",
        "test-frontend"
      ]
    },
    "enforce_admins": true,
    "required_pull_request_reviews": {
      "dismiss_stale_reviews": true,
      "require_code_owner_reviews": false,
      "required_approving_review_count": 1
    },
    "restrictions": null,
    "required_linear_history": true,
    "allow_force_pushes": false,
    "allow_deletions": false,
    "required_conversation_resolution": true
  }'

echo ""
echo "✅ Branch protection configured successfully"
echo "🔍 View settings: https://github.com/$OWNER/$REPO/settings/branches"
```

**Usage**:
```bash
export GITHUB_TOKEN="ghp_your_token_here"
chmod +x scripts/setup-branch-protection.sh
./scripts/setup-branch-protection.sh
```

---

## Configuration Files

### 1. CodeRabbit Configuration

**File**: `.coderabbit.yaml`

```yaml
# CodeRabbit Configuration for Rafiki
# Docs: https://docs.coderabbit.ai/configuration

language: "en-US"
early_access: false

reviews:
  profile: "assertive"  # Options: chill, assertive
  request_changes_workflow: false
  high_level_summary: true
  poem: false
  review_status: true
  collapse_walkthrough: false

  auto_review:
    enabled: true
    drafts: false
    base_branches:
      - main

  # Exclude files from review
  path_filters:
    - "!vendor/**"
    - "!frontend/node_modules/**"
    - "!frontend/.next/**"
    - "!**/*.pb.go"
    - "!**/*_generated.go"

  # Tool integrations
  tools:
    golangci-lint:
      enabled: true
      config_file: ".golangci.yml"

    eslint:
      enabled: true
      config_file: "frontend/eslint.config.mjs"

  # Path-specific instructions
  path_instructions:
    - path: "business/types/**"
      instructions: |
        CRITICAL: This directory contains business validation types with the Parse pattern.
        Never suggest auto-fixes that change validation logic.
        These types enforce domain rules and must be manually reviewed.

    - path: "business/domain/**/model.go"
      instructions: |
        IMPORTANT: These are domain models that define API contracts.
        Changes here affect frontend integration.
        Flag any modifications for manual review and frontend coordination.

    - path: "app/sdk/auth/**"
      instructions: |
        SECURITY SENSITIVE: Authentication and JWT handling.
        Never auto-fix. All changes require security review.

    - path: "foundation/keystore/**"
      instructions: |
        SECURITY SENSITIVE: Cryptographic key management.
        Never auto-fix. All changes require security review.

    - path: "frontend/lib/auth-context.tsx"
      instructions: |
        SECURITY SENSITIVE: Client-side authentication state.
        Changes affect security posture. Require manual review.

    - path: "frontend/lib/api.ts"
      instructions: |
        IMPORTANT: API client that integrates with Go backend.
        Changes must coordinate with backend API contracts.

    - path: "api/services/partners/mux/**"
      instructions: |
        Follow Go 1.22+ HTTP routing patterns.
        Use mux.HandleFunc with method patterns like "GET /endpoint".
        Prefer standard library over third-party routers.

    - path: "frontend/app/**"
      instructions: |
        Follow Next.js 16 App Router conventions.
        Prefer server components by default.
        Use 'use client' directive only when necessary.

chat:
  auto_reply: true

# Labels to add to PRs
labels:
  - "coderabbit"
```

### 2. Go Linter Configuration

**File**: `.golangci.yml`

```yaml
# golangci-lint configuration for Rafiki
# Docs: https://golangci-lint.run/usage/configuration/

run:
  timeout: 10m
  tests: true
  skip-dirs:
    - vendor
  modules-download-mode: readonly

linters-settings:
  gofmt:
    simplify: true

  gofumpt:
    module-path: github.com/francowini/rafiki
    extra-rules: true

  goimports:
    local-prefixes: github.com/francowini/rafiki

  gci:
    sections:
      - standard  # Standard library
      - default   # External packages
      - prefix(github.com/francowini/rafiki)  # Local packages
    skip-generated: true

  errcheck:
    check-type-assertions: true
    check-blank: true
    exclude-functions:
      - (io.Closer).Close
      - (*os.File).Close

  govet:
    enable-all: true
    disable:
      - shadow  # Too many false positives

  gosimple:
    checks: ["all"]

  staticcheck:
    checks: ["all"]

  unused:
    check-exported: false

  depguard:
    rules:
      main:
        deny:
          - pkg: io/ioutil
            desc: "Use io or os instead of deprecated io/ioutil"

linters:
  enable:
    # Auto-fixable linters (Tier 1)
    - gofmt
    - gofumpt
    - goimports
    - gci
    - gosimple
    - ineffassign
    - whitespace
    - canonicalheader

    # Important linters (no auto-fix)
    - errcheck
    - govet
    - staticcheck
    - typecheck
    - bodyclose
    - exportloopref
    - gocritic
    - revive
    - misspell
    - unconvert
    - unparam
    - nakedret

  disable:
    - exhaustruct  # Too strict
    - varnamelen   # Conflicts with idiomatic Go
    - gochecknoglobals  # We use globals appropriately
    - gochecknoinits    # We use init() appropriately

issues:
  exclude-rules:
    # Exclude some linters from test files
    - path: _test\.go
      linters:
        - errcheck
        - gosec
        - dupl

    # Exclude business types from unused checks (public API)
    - path: business/types/
      linters:
        - unused

    # Exclude vendor
    - path: vendor/
      linters:
        - all

  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0

output:
  formats:
    - format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
  sort-results: true
```

### 3. Prettier Configuration

**File**: `frontend/.prettierrc`

```json
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false,
  "arrowParens": "always",
  "endOfLine": "lf",
  "bracketSpacing": true,
  "jsxSingleQuote": false,
  "jsxBracketSameLine": false
}
```

**File**: `frontend/.prettierignore`

```
# Dependencies
node_modules
.pnp
.pnp.js

# Build output
.next
out
build
dist

# Cache
.turbo
.cache

# Environment
.env
.env.local
.env.production.local

# Config files (don't format)
package-lock.json
pnpm-lock.yaml
yarn.lock

# Generated
*.min.js
*.min.css
```

### 4. Enhanced ESLint Configuration

**File**: `frontend/eslint.config.mjs` (enhanced)

```javascript
import { defineConfig } from 'eslint/config';
import nextCoreWebVitals from 'eslint-config-next/core-web-vitals';
import nextTypescript from 'eslint-config-next/typescript';

const eslintConfig = defineConfig([
  ...nextCoreWebVitals,
  ...nextTypescript,
  {
    rules: {
      // TypeScript
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
      '@typescript-eslint/no-explicit-any': 'warn', // Warn but don't auto-fix

      // Code style (auto-fixable)
      'quotes': ['error', 'single', { avoidEscape: true, allowTemplateLiterals: true }],
      'semi': ['error', 'always'],
      'comma-dangle': ['error', 'always-multiline'],
      'indent': ['error', 2, { SwitchCase: 1 }],

      // Best practices
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      'prefer-const': 'error',
      'no-var': 'error',

      // React
      'react/jsx-curly-spacing': ['error', { when: 'never', children: true }],
      'react/jsx-boolean-value': ['error', 'never'],
      'react/self-closing-comp': 'error',

      // Imports (if using eslint-plugin-import)
      // 'import/order': ['error', {
      //   'groups': ['builtin', 'external', 'internal', 'parent', 'sibling', 'index'],
      //   'newlines-between': 'always',
      //   'alphabetize': { order: 'asc' }
      // }],
    },
  },
  {
    ignores: [
      '.next/**',
      'out/**',
      'build/**',
      'node_modules/**',
      'next-env.d.ts',
    ],
  },
]);

export default eslintConfig;
```

### 5. Frontend Package.json Scripts

**File**: `frontend/package.json` (scripts section)

```json
{
  "scripts": {
    "dev": "next dev --webpack",
    "build": "next build --webpack",
    "start": "next start",
    "lint": "eslint . --ext .ts,.tsx",
    "lint:fix": "eslint . --ext .ts,.tsx --fix",
    "format": "prettier --write \"**/*.{ts,tsx,json,md}\"",
    "format:check": "prettier --check \"**/*.{ts,tsx,json,md}\"",
    "typecheck": "tsc --noEmit",
    "check": "npm run typecheck && npm run lint && npm run format:check"
  }
}
```

---

## Workflow Examples

### Example 1: Normal PR Development Workflow

**Scenario**: Developer creates a PR with new feature

```bash
# 1. Developer creates feature branch
git checkout -b feature/add-habit-tracking
# ... make changes ...
git add .
git commit -m "feat: add habit tracking feature"

# Pre-commit hook runs automatically:
# ✅ Formats Go code with gofmt
# ✅ Organizes imports with goimports
# ✅ Formats TypeScript with Prettier
# ✅ Runs ESLint auto-fix

# 2. Push to GitHub
git push origin feature/add-habit-tracking

# 3. Create PR
gh pr create --title "Add habit tracking feature" --body "..."

# GitHub Actions run automatically:
# 🤖 pr-autofix.yml (Tier 1 auto-fixes)
# 🤖 pr-checks.yml (linting, tests, type checks)
# 🤖 CodeRabbit Pro reviews the PR

# 4. Developer reviews CodeRabbit comments
# 5. Developer runs Claude Code command
/coderabbit-review

# Claude Code command:
# - Fetches CodeRabbit comments
# - Applies Tier 1 fixes automatically
# - Presents Tier 2 interactively:
#   "Apply error wrapping fix? (y/n/s/v)"
# - Flags Tier 3 for manual review
# - Commits and pushes changes

# 6. CodeRabbit re-reviews
# 7. All status checks pass
# 8. Reviewer approves PR
# 9. Developer merges PR
```

### Example 2: Fixing CodeRabbit Issues Interactively

**Scenario**: PR has 15 CodeRabbit suggestions

```bash
# Developer runs command
/coderabbit-review

# Output:
📋 Found PR #42
🤖 Found 15 CodeRabbit suggestions

Categorizing suggestions:
- Tier 1 (auto-fix): 8 formatting issues
- Tier 2 (review): 5 logic improvements
- Tier 3 (manual): 2 security concerns

🔧 Applying Tier 1 safe auto-fixes...
  → Running Go formatters...
  → Running frontend formatters...
✅ Tier 1 fixes applied

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📍 Tier 2 Suggestion 1/5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

File: business/domain/momentbus/momentbus.go:45

💬 CodeRabbit suggests improving error wrapping:

BEFORE:
return fmt.Errorf("failed to create moment: %s", err)

AFTER:
return fmt.Errorf("failed to create moment: %w", err)

Reason: Use %w to wrap errors for better error inspection

Apply this fix? [y]es / [n]o / [s]kip remaining / [v]iew context
> y

✅ Applied

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📍 Tier 2 Suggestion 2/5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

File: frontend/components/MomentForm.tsx:124

💬 CodeRabbit suggests improving type safety:

BEFORE:
} catch (err: any) {
  setError(err.message || 'Error saving moment');
}

AFTER:
} catch (err: unknown) {
  const error = err as Error;
  setError(error.message || 'Error saving moment');
}

Reason: Avoid using 'any' type for better type safety

Apply this fix? [y]es / [n]o / [s]kip remaining / [v]iew context
> y

✅ Applied

# ... continues for all Tier 2 suggestions ...

⚠️  Manual Review Required (Tier 3)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The following issues require manual review:

- [ ] app/sdk/auth/auth.go:67
      Reason: Security-sensitive file (JWT token handling)
      Link: https://github.com/.../pull/42#discussion_r123456

- [ ] business/types/intensity/intensity.go:15
      Reason: Business validation logic
      Link: https://github.com/.../pull/42#discussion_r123457

💾 Committing changes...
📤 Pushing to branch feature/add-habit-tracking...
✅ Changes pushed to PR #42

📝 Summary:
- Tier 1 (auto): 8 issues fixed
- Tier 2 (reviewed): 5 issues fixed (all approved)
- Tier 3 (manual): 2 issues flagged for review
```

### Example 3: Tier 1 Only Quick Fix

**Scenario**: Developer just wants formatting fixes

```bash
/coderabbit-review --tier1-only

# Output:
📋 Found PR #42
🤖 Found 15 CodeRabbit suggestions

🔧 Tier 1 only mode - applying safe auto-fixes...
  → Running Go formatters...
  → Running frontend formatters...
✅ Tier 1 fixes applied (8 issues)

ℹ️  Skipped Tier 2/3 suggestions (use without --tier1-only to review)

💾 Committing changes...
📤 Pushing to branch feature/add-habit-tracking...
✅ Done!
```

---

## Dependencies and Coordination

### Installation Dependencies

#### Backend Tools
```bash
# Go toolchain
go install golang.org/x/tools/cmd/goimports@latest
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### Frontend Tools
```bash
cd frontend
npm install --save-dev prettier
# ESLint already installed
```

#### GitHub CLI (for slash command)
```bash
# macOS
brew install gh

# Authenticate
gh auth login
```

### Team Coordination Requirements

#### One-Time Setup (per developer)

1. **Install local tools**:
   ```bash
   make setup-tools  # New Makefile target to install all tools
   ```

2. **Install git hooks**:
   ```bash
   ./scripts/install-git-hooks.sh
   ```

3. **Authenticate GitHub CLI**:
   ```bash
   gh auth login
   ```

#### Team Communication

**Before Rollout**:
- [ ] Team meeting to explain new workflow
- [ ] Demo of `/coderabbit-review` command
- [ ] Document in team wiki/Notion
- [ ] Pair on first few PRs with new workflow

**Documentation Needed**:
- Update [README.md](README.md) with new workflow
- Create `CONTRIBUTING.md` with code quality guidelines
- Add section to deployment docs about pre-commit hooks

### Dependency Matrix

| Component | Depends On | Installation |
|-----------|-----------|--------------|
| CodeRabbit Pro | GitHub App | One-time install (done ✅) |
| Branch Protection | GitHub settings | One-time config |
| Pre-commit hooks | Go tools, npm | Per-developer setup |
| GitHub Actions | Blacksmith app | One-time GitHub App install |
| `/coderabbit-review` | GitHub CLI, git | Per-developer setup |
| Prettier | npm | `npm install` |
| ESLint | npm | Already installed ✅ |
| golangci-lint | Go | `go install` |

### API Rate Limits

**GitHub API**:
- **Authenticated**: 5,000 requests/hour
- **/coderabbit-review command** uses ~10-15 requests per execution
- **Mitigation**: Cache PR data, batch operations

**CodeRabbit**:
- No explicit API (uses GitHub comments)
- No rate limiting concerns

---

## Deployment Strategy

### Phase 1: Foundation Setup (Week 1)

**Goal**: Set up infrastructure without disrupting current workflow

**Tasks**:
1. ✅ Create `.coderabbit.yaml` configuration
2. ✅ Create `.golangci.yml` configuration
3. ✅ Create `frontend/.prettierrc` and run initial formatting
4. ✅ Create enhanced `frontend/eslint.config.mjs`
5. ✅ Update `frontend/package.json` scripts
6. ✅ Configure GitHub branch protection
7. ✅ Test configurations on a throwaway branch

**Success Criteria**:
- All config files committed to main
- Branch protection active
- CodeRabbit reviewing PRs (already active)
- No disruption to current development

**Rollback Plan**: Delete config files, disable branch protection

### Phase 2: GitHub Actions (Week 2)

**Goal**: Automate Tier 1 fixes in CI/CD

**Tasks**:
1. ✅ Install Blacksmith GitHub App
2. ✅ Create `pr-autofix.yml` workflow
3. ✅ Create `pr-checks.yml` workflow
4. ✅ Test workflows on a test PR
5. ✅ Monitor workflow performance and adjust timeouts
6. ✅ Add workflow badges to README

**Success Criteria**:
- Workflows run on every PR
- Auto-fixes committed automatically
- Status checks required for merge
- Workflows complete in <5 minutes

**Rollback Plan**: Disable workflows, remove from branch protection

### Phase 3: Pre-Commit Hooks (Week 3)

**Goal**: Enable local auto-fixing

**Tasks**:
1. ✅ Create `scripts/install-git-hooks.sh`
2. ✅ Create `scripts/setup-tools.sh` (install all Go/npm tools)
3. ✅ Document in CONTRIBUTING.md
4. ✅ Team training session
5. ✅ Each developer installs hooks
6. ✅ Monitor for issues and adjust hook scripts

**Success Criteria**:
- All team members have hooks installed
- Hooks run in <5 seconds for typical commits
- No false positives blocking commits
- Developer satisfaction >80%

**Rollback Plan**: Developers can bypass with `--no-verify`

### Phase 4: Claude Code Command (Week 4)

**Goal**: Enable interactive Tier 2/3 review

**Tasks**:
1. ✅ Create `.claude/commands/coderabbit-review.md`
2. ✅ Implement categorization logic
3. ✅ Implement interactive prompts
4. ✅ Test on multiple PRs with various issue types
5. ✅ Document command usage
6. ✅ Team demo and training

**Success Criteria**:
- Command successfully fetches and categorizes comments
- Interactive prompts work correctly
- Tier 1 fixes applied automatically
- Tier 2 approval workflow intuitive
- Tier 3 issues flagged appropriately

**Rollback Plan**: Command is optional, no rollback needed

### Phase 5: Optimization (Ongoing)

**Goal**: Refine based on real-world usage

**Tasks**:
- Monitor categorization accuracy
- Adjust Tier 1/2/3 rules based on false positives
- Optimize workflow performance
- Add more path-specific CodeRabbit instructions
- Gather team feedback and iterate

---

## Risk Analysis and Mitigation

### Risk 1: Over-Automation Breaking Code

**Risk**: Auto-fixes introduce bugs or break functionality

**Likelihood**: Low (Tier 1 is formatting only)
**Impact**: Medium (would require fix and re-deploy)

**Mitigation**:
- Tier 1 limited to formatting/style only (no logic changes)
- All fixes run through required status checks (tests must pass)
- Developer reviews all auto-committed changes in PR
- Easy rollback via git

**Detection**:
- PR checks fail after auto-fix
- Developer notices unexpected changes in diff

**Response**:
- Immediately disable problematic auto-fix rule
- Revert auto-fix commit
- Manually fix the issue
- Update categorization logic

### Risk 2: GitHub Actions Costs

**Risk**: Blacksmith CI costs more than expected

**Likelihood**: Low
**Impact**: Low (monitored monthly)

**Mitigation**:
- Start with 2vcpu runners (cost-effective)
- Monitor usage in first month
- Set up billing alerts
- Can fall back to free GitHub runners if needed

**Detection**: GitHub billing dashboard

**Response**:
- Switch to standard runners: `runs-on: ubuntu-latest`
- Optimize workflow (cache dependencies, run only on changed files)

### Risk 3: Pre-Commit Hook Frustration

**Risk**: Developers find hooks slow or annoying

**Likelihood**: Medium
**Impact**: Low (optional feature)

**Mitigation**:
- Hooks only run on staged files (fast)
- Clear instructions for bypass: `git commit --no-verify`
- Hooks are optional (CI still catches issues)
- Gather feedback and adjust

**Detection**: Developer complaints, frequent use of `--no-verify`

**Response**:
- Optimize hook performance
- Reduce scope of checks
- Make hooks optional
- Improve developer documentation

### Risk 4: False Positive Tier Categorization

**Risk**: Command miscategorizes issues (e.g., Tier 1 should be Tier 2)

**Likelihood**: Medium (categorization is heuristic)
**Impact**: Low (developer can skip or manually review)

**Mitigation**:
- Conservative categorization (when in doubt, use Tier 2)
- Developer can always skip Tier 2 suggestions
- Tier 3 explicit list of protected paths
- Regular review and tuning of rules

**Detection**: Developer reports, monitoring of applied fixes

**Response**:
- Update categorization logic in command
- Add more explicit path rules
- Improve CodeRabbit instructions

### Risk 5: Merge Conflicts from Auto-Commits

**Risk**: Auto-fix commits conflict with developer's work

**Likelihood**: Low
**Impact**: Low (standard git conflict resolution)

**Mitigation**:
- Workflow concurrency group prevents parallel runs
- Pre-commit hooks apply fixes before push
- Clear commit messages identify auto-fixes
- Developer pulls before pushing

**Detection**: Git merge conflict errors

**Response**:
- Standard git conflict resolution
- `git pull --rebase` to apply fixes on top

### Risk 6: GitHub API Rate Limiting

**Risk**: `/coderabbit-review` command hits API limits

**Likelihood**: Very Low (5,000 requests/hour)
**Impact**: Low (command fails gracefully)

**Mitigation**:
- Command uses ~10-15 requests per run
- GitHub CLI handles rate limiting automatically
- Clear error messages
- Caching where possible

**Detection**: Command shows rate limit error

**Response**:
- Wait for rate limit reset (shown in error)
- Use PAT token with higher limits if needed

### Risk 7: Breaking Changes to CodeRabbit API/Format

**Risk**: CodeRabbit changes comment format, breaking command

**Likelihood**: Low (stable API)
**Impact**: Medium (command stops working)

**Mitigation**:
- Use official GitHub API (stable)
- Graceful error handling
- Regular testing
- Monitor CodeRabbit changelog

**Detection**: Command fails to parse comments

**Response**:
- Update parsing logic
- Add fallback patterns
- Contact CodeRabbit support

---

## Success Metrics

### Quantitative Metrics

**Code Quality**:
- ✅ **Formatting issues**: Reduce by 90% (auto-fixed by Tier 1)
- ✅ **Linting issues**: Reduce by 70% (auto-fixed + enforced)
- ✅ **Test coverage**: Maintain >80% (enforced by CI)
- ✅ **CodeRabbit suggestions per PR**: Track over time (expect reduction)

**Developer Productivity**:
- ✅ **Time spent on formatting**: Reduce by 80% (automated)
- ✅ **PR review time**: Reduce by 30% (fewer style comments)
- ✅ **Time to merge**: Reduce by 20% (faster iterations)
- ✅ **Build/lint duration**: <5 minutes (Blacksmith speed)

**Automation**:
- ✅ **Tier 1 auto-fix success rate**: >95%
- ✅ **Tier 2 categorization accuracy**: >85%
- ✅ **Pre-commit hook adoption**: >90% of team
- ✅ **CI/CD success rate**: >95%

### Qualitative Metrics

**Developer Experience**:
- ✅ **Frustration with formatting**: Eliminated
- ✅ **Confidence in auto-fixes**: High
- ✅ **Understanding of code quality**: Improved (learning from CodeRabbit)
- ✅ **Satisfaction with workflow**: >8/10

**Code Review Quality**:
- ✅ **Focus on logic vs style**: More logic discussion
- ✅ **CodeRabbit suggestions acceptance**: >70%
- ✅ **Reviewer burden**: Reduced
- ✅ **PR quality**: Improved

### Tracking Method

**Weekly Dashboard** (to be created):
```markdown
## Week of 2025-11-17

### Automation
- PRs opened: 12
- Tier 1 auto-fixes applied: 8/12 (67%)
- Tier 2 suggestions reviewed: 23
- Tier 2 accepted: 18/23 (78%)
- Tier 3 manual reviews: 5

### Performance
- Average PR review time: 2.1 hours (↓ from 2.8 hours)
- Average time to merge: 1.3 days (↓ from 1.7 days)
- CI/CD duration: 4.2 minutes (target: <5 min) ✅

### Issues
- False positive categorizations: 2
- Pre-commit hook bypasses: 3
- CI failures due to auto-fix: 0 ✅
```

---

## Implementation Roadmap

### Immediate Actions (Week 1: Nov 18-22)

**Priority 1: Foundation**
- [ ] Create `.coderabbit.yaml` (2 hours)
- [ ] Create `.golangci.yml` (1 hour)
- [ ] Create `frontend/.prettierrc` (1 hour)
- [ ] **Run initial Prettier formatting** (1 hour + review)
- [ ] Update `frontend/package.json` scripts (30 min)
- [ ] Update `frontend/eslint.config.mjs` (1 hour)

**Priority 2: Branch Protection**
- [ ] Create `scripts/setup-branch-protection.sh` (1 hour)
- [ ] Run script to configure GitHub (30 min)
- [ ] Test branch protection (30 min)

**Priority 3: Documentation**
- [✅] Create this implementation plan
- [ ] Update README.md with new workflow (1 hour)
- [ ] Create CONTRIBUTING.md (2 hours)

**Total Estimated Time**: ~11 hours
**Deliverable**: All configurations in place, branch protection active

### Week 2: GitHub Actions (Nov 25-29)

**Priority 1: Workflows**
- [ ] Install Blacksmith GitHub App (15 min)
- [ ] Create `.github/workflows/pr-autofix.yml` (2 hours)
- [ ] Create `.github/workflows/pr-checks.yml` (2 hours)
- [ ] Test workflows on test branch (1 hour)

**Priority 2: Optimization**
- [ ] Monitor workflow performance (ongoing)
- [ ] Adjust timeouts and caching (1 hour)
- [ ] Add workflow status badges to README (30 min)

**Priority 3: Status Checks**
- [ ] Update branch protection with required checks (30 min)
- [ ] Test PR merge with all checks (30 min)

**Total Estimated Time**: ~7.5 hours
**Deliverable**: Automated Tier 1 fixes in CI/CD

### Week 3: Pre-Commit Hooks (Dec 2-6)

**Priority 1: Scripts**
- [ ] Create `scripts/install-git-hooks.sh` (2 hours)
- [ ] Create `scripts/setup-tools.sh` (1 hour)
- [ ] Create `.git/hooks/pre-commit` template (2 hours)
- [ ] Test hooks locally (1 hour)

**Priority 2: Documentation**
- [ ] Document hook installation in CONTRIBUTING.md (1 hour)
- [ ] Create troubleshooting guide (1 hour)

**Priority 3: Team Rollout**
- [ ] Team training session (1 hour)
- [ ] Pair with each developer for setup (2 hours)
- [ ] Gather feedback and adjust (ongoing)

**Total Estimated Time**: ~11 hours
**Deliverable**: All developers have pre-commit hooks

### Week 4: Claude Code Command (Dec 9-13)

**Priority 1: Command Implementation**
- [ ] Create `.claude/commands/coderabbit-review.md` (4 hours)
- [ ] Implement categorization logic (3 hours)
- [ ] Implement interactive prompts (2 hours)
- [ ] Test on multiple PRs (2 hours)

**Priority 2: Edge Cases**
- [ ] Handle no PR found (1 hour)
- [ ] Handle no CodeRabbit comments (1 hour)
- [ ] Handle API rate limits (1 hour)
- [ ] Handle git conflicts (1 hour)

**Priority 3: Documentation & Training**
- [ ] Document command usage (1 hour)
- [ ] Create command demo video (1 hour)
- [ ] Team training session (1 hour)

**Total Estimated Time**: ~18 hours
**Deliverable**: `/coderabbit-review` command working end-to-end

### Week 5+: Monitoring & Optimization (Ongoing)

**Continuous Improvement**:
- [ ] Weekly metrics review
- [ ] Adjust Tier 1/2/3 rules based on false positives
- [ ] Optimize workflow performance
- [ ] Add more CodeRabbit path instructions
- [ ] Gather team feedback
- [ ] Iterate on command features

**Total Estimated Time**: ~2 hours/week ongoing
**Deliverable**: Continuously improving automation

---

## Total Implementation Estimate

**Total Development Time**: ~47.5 hours (~1.5 weeks of focused work)
**Calendar Time**: 4-5 weeks (accounting for testing, rollout, training)
**Team Impact**: ~1 hour per developer for setup and training
**Ongoing Maintenance**: ~2 hours/week for monitoring and optimization

---

## Conclusion

This implementation plan provides a comprehensive roadmap for integrating CodeRabbit automation into the Rafiki development workflow. The three-tier strategy balances automation efficiency with code safety:

- **Tier 1 (80% of issues)**: Fully automated via pre-commit hooks and CI/CD
- **Tier 2 (15% of issues)**: Interactive review via Claude Code command
- **Tier 3 (5% of issues)**: Manual review for security and business logic

**Key Benefits**:
- ⏱️ **Time Savings**: ~10-15 hours/month on formatting and code review
- 🎯 **Focus Shift**: From style to logic in code reviews
- 🛡️ **Safety**: Multi-layer enforcement (local hooks + CI + branch protection)
- 🤖 **Automation**: 80%+ of CodeRabbit suggestions auto-addressed
- 📈 **Quality**: Consistent code style and best practices enforcement

**Next Steps**:
1. Review and approve this plan
2. Begin Week 1 foundation work
3. Iterate based on real-world usage
4. Celebrate wins! 🎉

---

**Document Version**: 1.0
**Last Updated**: 2025-11-17
**Authors**: Backend Engineer + Frontend Engineer + DevOps Engineer
**Status**: Awaiting Approval ⏳
