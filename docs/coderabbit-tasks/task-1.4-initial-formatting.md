# Task 1.4: Run Initial Prettier Formatting

**Phase**: 1 - Foundation Setup
**Estimated Time**: 1 hour + review
**Priority**: High
**Dependencies**: Task 1.3 (Prettier config must exist)

## Objective

Run Prettier on the entire frontend codebase for the first time to establish a consistent formatting baseline. This will create a large formatting commit that should be reviewed carefully.

## Steps

1. **Create a dedicated branch**
   ```bash
   git checkout -b chore/initial-prettier-formatting
   ```

2. **Verify Prettier config exists**
   ```bash
   ls -la frontend/.prettierrc
   ls -la frontend/.prettierignore
   ```

3. **Run Prettier check first (dry run)**
   ```bash
   cd frontend
   npx prettier --check "**/*.{ts,tsx,json,md}"
   ```
   This shows what files will be changed without modifying them.

4. **Review the list of files to be formatted**
   - Ensure no unexpected files are included
   - Verify ignored paths are properly excluded

5. **Run Prettier write (actual formatting)**
   ```bash
   npx prettier --write "**/*.{ts,tsx,json,md}"
   cd ..
   ```

6. **Review the changes**
   ```bash
   git diff --stat
   git diff
   ```
   Look for:
   - Only formatting changes (no logic changes)
   - Consistent quote style (single quotes)
   - Consistent indentation (2 spaces)
   - Proper semicolon usage
   - No accidental changes to functionality

7. **Run type check to ensure no breakage**
   ```bash
   cd frontend
   npm run typecheck
   ```

8. **Run build to ensure no breakage**
   ```bash
   npm run build
   ```

9. **Commit the formatting changes**
   ```bash
   cd ..
   git add frontend/
   git commit -m "style: initial Prettier formatting for frontend

   Apply Prettier formatting to entire frontend codebase:
   - Enforces 2-space indentation
   - Single quotes for strings
   - Semicolons enabled
   - 100-character line width
   - LF line endings

   Part of CodeRabbit automation setup (Phase 1).
   Future formatting will be enforced via pre-commit hooks and CI.

   No functional changes - formatting only."
   ```

10. **Push and create PR**
    ```bash
    git push origin chore/initial-prettier-formatting
    gh pr create --title "Initial Prettier Formatting" \
      --body "Applies Prettier formatting to frontend codebase. See commit message for details. **Review carefully for any unintended changes.**"
    ```

## Expected Output

- Large commit with formatting changes across frontend
- All TypeScript, TSX, JSON, and Markdown files formatted consistently
- No functional changes (tests still pass)
- Build succeeds without errors

## Success Criteria

- [ ] Dedicated branch created for formatting changes
- [ ] Prettier applied to all frontend files
- [ ] Git diff shows only formatting changes (no logic changes)
- [ ] TypeScript compilation succeeds (`npm run typecheck`)
- [ ] Build succeeds (`npm run build`)
- [ ] Changes committed with descriptive commit message
- [ ] PR created for team review
- [ ] After approval, merged to main

## Reference

Main plan section: "Deployment Strategy → Phase 1: Foundation Setup"

## Testing Checklist

Before committing:
- [ ] No files from `.prettierignore` were modified
- [ ] No `node_modules` or `.next` files in diff
- [ ] Type checking passes
- [ ] Build passes
- [ ] No unexpected file changes

## Notes

**IMPORTANT**: This is a one-time operation that will touch many files. Review carefully.

**Common Issues**:
- If build fails, check for parsing errors in JSX/TSX
- If types fail, Prettier might have broken multi-line type definitions
- If unexpected files changed, update `.prettierignore`

**After Merge**:
- All future formatting will be automatic via pre-commit hooks (Phase 3)
- CI will enforce formatting on PRs (Phase 2)
- Team should pull main and rebase any open PRs to avoid conflicts
