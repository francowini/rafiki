# Task 3.3: Create Pre-Commit Hook Template

**Phase**: 3 - Pre-Commit Hooks
**Estimated Time**: 2 hours
**Priority**: High
**Dependencies**: Tasks 3.1, 3.2 (setup scripts exist)

## Objective

Create a pre-commit hook template that auto-formats code before commits. The hook should detect changed files, run appropriate formatters for Go and TypeScript, and fail fast on linting errors.

## Steps

1. **Create the pre-commit hook template**

   The template will be used by the installation script (Task 3.1) to create `.git/hooks/pre-commit` in each developer's local repository.

2. **Create template file**
   ```bash
   mkdir -p scripts/hooks
   touch scripts/hooks/pre-commit.template
   chmod +x scripts/hooks/pre-commit.template
   ```

3. **Add hook logic**

   The hook should:
   - Get list of staged files
   - Detect Go files (`.go`)
   - Detect TypeScript files (`.ts`, `.tsx` in `frontend/`)
   - Run formatters on changed files only (fast)
   - Re-stage formatted files
   - Run quick lint checks
   - Fail commit if linting fails (with bypass option)

4. **Implement file detection**
   ```bash
   STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)
   STAGED_TS_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(ts|tsx)$' | grep '^frontend/' || true)
   ```

5. **Implement Go formatting**
   ```bash
   if [ -n "$STAGED_GO_FILES" ]; then
     echo "  → Formatting Go files..."
     echo "$STAGED_GO_FILES" | xargs gofmt -w -s
     echo "$STAGED_GO_FILES" | xargs goimports -w -local github.com/francowini/rafiki
     echo "$STAGED_GO_FILES" | xargs git add
   fi
   ```

6. **Implement TypeScript formatting**
   ```bash
   if [ -n "$STAGED_TS_FILES" ]; then
     echo "  → Formatting TypeScript files..."
     cd frontend
     echo "$STAGED_TS_FILES" | sed 's|^frontend/||' | xargs npx prettier --write
     echo "$STAGED_TS_FILES" | sed 's|^frontend/||' | xargs npx eslint --fix || true
     cd ..
     echo "$STAGED_TS_FILES" | xargs git add
   fi
   ```

7. **Add quick lint checks**
   ```bash
   # Go: Quick lint on changed files only
   if [ -n "$STAGED_GO_FILES" ]; then
     golangci-lint run --fast --new-from-rev=HEAD~1 || {
       echo "❌ Linting failed. Fix issues or use 'git commit --no-verify'"
       exit 1
     }
   fi

   # TypeScript: Quick type check
   if [ -n "$STAGED_TS_FILES" ]; then
     cd frontend
     npx tsc --noEmit || {
       echo "❌ Type check failed. Fix issues or use 'git commit --no-verify'"
       exit 1
     }
     cd ..
   fi
   ```

8. **Add helpful output**
   - Show what's being formatted
   - Show progress indicators
   - Show success/failure clearly
   - Show bypass option in error messages

9. **Test the hook**
   ```bash
   # Manually copy to .git/hooks for testing
   cp scripts/hooks/pre-commit.template .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit

   # Test with unformatted Go file
   echo 'package main; func main(){println("test")}' > test.go
   git add test.go
   git commit -m "test"
   # Should auto-format and commit

   # Cleanup
   git reset HEAD~1
   rm test.go
   ```

10. **Commit the template**
    ```bash
    git add scripts/hooks/pre-commit.template
    git commit -m "chore: add pre-commit hook template

    Add pre-commit hook template that:
    - Auto-formats Go code (gofmt, goimports)
    - Auto-formats TypeScript (Prettier, ESLint)
    - Runs only on staged files (fast)
    - Quick lint checks before commit
    - Fail-fast with bypass option (--no-verify)
    - Clear output and error messages

    Template is installed by scripts/install-git-hooks.sh
    Part of CodeRabbit automation (Phase 3)."
    ```

## Expected Output

File: `scripts/hooks/pre-commit.template`

Executable shell script that:
- Detects staged files
- Formats Go and TypeScript separately
- Re-stages formatted files
- Runs quick lints
- Provides clear output

## Success Criteria

- [ ] Pre-commit hook template created
- [ ] Executable permissions set
- [ ] Detects Go files correctly
- [ ] Detects TypeScript files correctly
- [ ] Formats Go files (gofmt, goimports)
- [ ] Formats TypeScript files (Prettier, ESLint)
- [ ] Re-stages formatted files automatically
- [ ] Quick lint checks run before commit
- [ ] Fails gracefully with clear error messages
- [ ] Bypass option documented (--no-verify)
- [ ] Template committed to repository

## Reference

Main plan section: "Pre-Commit Hooks Strategy → Option 2: Custom Git Hook Script"

## Testing Scenarios

### Test 1: Go Formatting
```bash
echo 'package main;func test(){return}' > test.go
git add test.go
git commit -m "test"
# Expected: Auto-formatted with proper spacing
```

### Test 2: TypeScript Formatting
```bash
cd frontend
echo 'const x={a:1}' > test.ts
git add test.ts
git commit -m "test"
# Expected: Auto-formatted with quotes, spacing, semicolon
```

### Test 3: Lint Failure
```bash
echo 'package main; var unused int' > test.go
git add test.go
git commit -m "test"
# Expected: Lint error, commit blocked
```

### Test 4: Bypass Hook
```bash
git commit --no-verify -m "test"
# Expected: Commit succeeds, hook skipped
```

## Notes

**Performance**:
- Hook only runs on staged files (not entire codebase)
- Should complete in <5 seconds for typical commits
- Larger commits may take longer (acceptable)

**Bypass Option**:
- Always available via `git commit --no-verify`
- Use for WIP commits or emergencies
- CI will still catch issues

**Error Messages**:
- Should be clear and actionable
- Include bypass option in error output
- Don't be too verbose (quick feedback)

**Integration**:
- Template is copied to `.git/hooks/pre-commit` by install script
- Each developer runs install script once
- Hook is local (not committed to repo)
