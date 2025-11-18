# Task 2.2: Create PR Auto-Fix Workflow

**Phase**: 2 - GitHub Actions
**Estimated Time**: 2 hours
**Priority**: High
**Dependencies**: Task 2.1 (Blacksmith CI installed)

## Objective

Create a GitHub Actions workflow (`.github/workflows/pr-autofix.yml`) that automatically applies Tier 1 formatting fixes when a PR is opened or updated. The workflow runs on Blacksmith CI for faster execution.

## Steps

1. **Create workflows directory**
   ```bash
   mkdir -p .github/workflows
   ```

2. **Create the workflow file**
   ```bash
   touch .github/workflows/pr-autofix.yml
   ```

3. **Add workflow configuration**

   Key components:
   - **Trigger**: On PR open, sync, reopen to `main`
   - **Permissions**: `contents: write`, `pull-requests: write`
   - **Concurrency**: Group by PR, cancel in-progress
   - **Jobs**:
     - `autofix-backend`: Go formatting (Tier 1 only)
     - `autofix-frontend`: Frontend formatting (Tier 1 only)

4. **Backend auto-fix job**
   - Runs on: `blacksmith-2vcpu-ubuntu-2204`
   - Condition: Only if `.go` files changed
   - Steps:
     - Checkout PR branch
     - Setup Go 1.25.3
     - Install goimports, golangci-lint
     - Run gofmt, goimports, golangci-lint --fix (Tier 1 linters only)
     - Commit and push if changes detected

5. **Frontend auto-fix job**
   - Runs on: `blacksmith-2vcpu-ubuntu-2204`
   - Condition: Only if `frontend/` files changed
   - Steps:
     - Checkout PR branch
     - Setup Node.js 20
     - Install dependencies
     - Run Prettier and ESLint --fix
     - Commit and push if changes detected

6. **Test the workflow**
   - Create a test PR with unformatted code
   - Verify workflow triggers
   - Verify auto-fixes are committed
   - Verify only Tier 1 changes (no logic changes)

7. **Commit the workflow**
   ```bash
   git add .github/workflows/pr-autofix.yml
   git commit -m "ci: add PR auto-fix workflow (Tier 1)

   Add GitHub Actions workflow for automatic code formatting:
   - Runs on Blacksmith CI (faster runners)
   - Backend: gofmt, goimports, golangci-lint safe fixes
   - Frontend: Prettier, ESLint safe fixes
   - Only applies Tier 1 (formatting) changes
   - Runs in parallel for backend and frontend
   - Auto-commits fixes to PR branch

   Part of CodeRabbit automation (Phase 2)."
   ```

## Expected Output

File: `.github/workflows/pr-autofix.yml`

Workflow with:
- Trigger on PR events to main
- Two parallel jobs (backend, frontend)
- Blacksmith CI runners
- Auto-commit of Tier 1 fixes
- Proper concurrency control

## Success Criteria

- [ ] Workflow file created in `.github/workflows/`
- [ ] Triggers on PR open/sync/reopen
- [ ] Uses Blacksmith CI runners
- [ ] Backend job runs only on Go file changes
- [ ] Frontend job runs only on frontend file changes
- [ ] Only Tier 1 safe fixes applied
- [ ] Changes auto-committed with clear message
- [ ] Jobs run in parallel
- [ ] Concurrency group prevents duplicate runs
- [ ] Test PR shows workflow working
- [ ] Changes committed to repository

## Reference

Main plan section: "GitHub Actions with Blacksmith CI → Workflow 1: PR Auto-Fix (Tier 1)"

## Testing Checklist

Create test PR with these issues:
- [ ] Unformatted Go code (wrong indentation)
- [ ] Missing imports in Go
- [ ] Unformatted TypeScript (wrong quotes, semicolons)
- [ ] Verify workflow commits fixes automatically
- [ ] Verify commit message is clear
- [ ] Verify no logic changes in auto-fixes

## Sample Workflow Structure

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
    # ... steps

  autofix-frontend:
    name: Auto-fix Frontend Code
    runs-on: blacksmith-2vcpu-ubuntu-2204
    # ... steps
```

## Notes

**IMPORTANT Tier 1 Rules**:
- Only formatting/style changes
- No logic changes
- No changes to business types
- No changes to security-sensitive code

**Commit Message Format**:
```
style: auto-fix [backend|frontend] formatting (Tier 1)

🤖 Automated fixes by GitHub Actions:
- [list of tools run]
```

**Common Issues**:
- If workflow doesn't trigger: Check PR target branch is `main`
- If auto-commit fails: Check permissions in workflow
- If wrong files fixed: Check path conditionals in `if:` statements
