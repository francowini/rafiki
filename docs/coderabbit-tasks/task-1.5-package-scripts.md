# Task 1.5: Update Frontend Package Scripts

**Phase**: 1 - Foundation Setup
**Estimated Time**: 30 minutes
**Priority**: Medium
**Dependencies**: Task 1.3 (Prettier must be installed)

## Objective

Add npm scripts to `frontend/package.json` for running formatting, linting, and type checking consistently across development, CI, and pre-commit hooks.

## Steps

1. **Open the package.json file**
   ```bash
   code frontend/package.json
   # or your preferred editor
   ```

2. **Add/Update scripts section**
   Add the following scripts to the `scripts` section:

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

3. **Verify existing scripts**
   - Keep existing `dev`, `build`, `start` scripts as-is
   - Merge with new linting/formatting scripts
   - Don't duplicate existing scripts

4. **Test each new script**
   ```bash
   cd frontend

   # Test linting
   npm run lint

   # Test auto-fix
   npm run lint:fix

   # Test formatting
   npm run format:check

   # Test type checking
   npm run typecheck

   # Test combined check
   npm run check
   ```

5. **Commit the changes**
   ```bash
   git add frontend/package.json
   git commit -m "chore: add npm scripts for formatting and linting

   Add scripts for:
   - lint: ESLint check
   - lint:fix: ESLint auto-fix
   - format: Prettier write
   - format:check: Prettier check (CI)
   - typecheck: TypeScript compilation check
   - check: Run all checks (lint + format + types)

   These scripts will be used in:
   - Pre-commit hooks (Phase 3)
   - GitHub Actions CI (Phase 2)
   - Developer workflows"
   ```

## Expected Output

Updated `frontend/package.json` with new scripts:

```json
{
  "scripts": {
    "lint": "eslint . --ext .ts,.tsx",
    "lint:fix": "eslint . --ext .ts,.tsx --fix",
    "format": "prettier --write \"**/*.{ts,tsx,json,md}\"",
    "format:check": "prettier --check \"**/*.{ts,tsx,json,md}\"",
    "typecheck": "tsc --noEmit",
    "check": "npm run typecheck && npm run lint && npm run format:check"
  }
}
```

## Success Criteria

- [ ] All new scripts added to `package.json`
- [ ] `npm run lint` executes without errors
- [ ] `npm run lint:fix` applies fixes successfully
- [ ] `npm run format` formats files correctly
- [ ] `npm run format:check` validates formatting
- [ ] `npm run typecheck` compiles TypeScript
- [ ] `npm run check` runs all checks in sequence
- [ ] Changes committed to repository

## Reference

Main plan section: "Configuration Files → 5. Frontend Package.json Scripts"

## Testing

Run each script and verify expected behavior:

```bash
cd frontend

# Should show any ESLint errors/warnings
npm run lint

# Should fix auto-fixable issues
npm run lint:fix

# Should format all files
npm run format

# Should check formatting without changing files
npm run format:check

# Should check TypeScript types
npm run typecheck

# Should run all checks (use in CI)
npm run check
```

## Notes

- `format:check` is for CI (fails on unformatted files)
- `format` is for local dev (actually formats files)
- `lint:fix` is for pre-commit hooks (auto-fixes safe issues)
- `check` combines all checks for CI/pre-commit validation
- These scripts will be referenced in GitHub Actions workflows (Phase 2)
