# Task 1.6: Update ESLint Configuration

**Phase**: 1 - Foundation Setup
**Estimated Time**: 1 hour
**Priority**: Medium
**Dependencies**: None

## Objective

Enhance the existing `frontend/eslint.config.mjs` with additional rules for code style, TypeScript, and React best practices. Configure which rules can auto-fix (Tier 1) vs. require manual review.

## Steps

1. **Backup existing config**
   ```bash
   cp frontend/eslint.config.mjs frontend/eslint.config.mjs.backup
   ```

2. **Review current configuration**
   ```bash
   cat frontend/eslint.config.mjs
   ```
   Note existing rules and plugins.

3. **Add/Update rules section**
   Update the config with these key rules:

   **TypeScript rules:**
   - `@typescript-eslint/no-unused-vars`: Error (with _ prefix ignore)
   - `@typescript-eslint/no-explicit-any`: Warn (don't auto-fix)

   **Code style (auto-fixable):**
   - `quotes`: Single quotes with template literal allowance
   - `semi`: Always require semicolons
   - `comma-dangle`: Always for multiline
   - `indent`: 2 spaces with switch case handling

   **Best practices:**
   - `no-console`: Warn (allow warn/error)
   - `no-debugger`: Error
   - `prefer-const`: Error
   - `no-var`: Error

   **React rules:**
   - `react/jsx-curly-spacing`: Never
   - `react/jsx-boolean-value`: Never
   - `react/self-closing-comp`: Error

4. **Configure ignores**
   Ensure these paths are ignored:
   - `.next/**`
   - `out/**`
   - `build/**`
   - `node_modules/**`
   - `next-env.d.ts`

5. **Test the configuration**
   ```bash
   cd frontend
   npm run lint
   ```

6. **Test auto-fix**
   ```bash
   npm run lint:fix
   ```

7. **Verify TypeScript compatibility**
   ```bash
   npm run typecheck
   ```

8. **Commit the changes**
   ```bash
   git add frontend/eslint.config.mjs
   git commit -m "chore: enhance ESLint configuration

   Add rules for:
   - TypeScript best practices
   - Code style consistency (auto-fixable)
   - React/JSX patterns
   - Proper ignores for generated files

   Rules categorized:
   - Tier 1 (auto-fix): quotes, semi, indent, imports
   - Tier 2 (warn): no-explicit-any, console
   - Tier 3 (error): no-debugger, prefer-const

   Supports CodeRabbit automation (Phase 1)."
   ```

## Expected Output

Enhanced `frontend/eslint.config.mjs` with:

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
      '@typescript-eslint/no-unused-vars': ['error', {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
        caughtErrorsIgnorePattern: '^_',
      }],
      '@typescript-eslint/no-explicit-any': 'warn',

      // Code style (auto-fixable)
      'quotes': ['error', 'single', { avoidEscape: true }],
      'semi': ['error', 'always'],
      'comma-dangle': ['error', 'always-multiline'],
      'indent': ['error', 2, { SwitchCase: 1 }],

      // Best practices
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      'prefer-const': 'error',
      'no-var': 'error',

      // React
      'react/jsx-curly-spacing': ['error', { when: 'never' }],
      'react/jsx-boolean-value': ['error', 'never'],
      'react/self-closing-comp': 'error',
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

## Success Criteria

- [ ] ESLint config enhanced with new rules
- [ ] Auto-fixable rules clearly identified
- [ ] TypeScript-specific rules configured
- [ ] React/JSX rules added
- [ ] Proper ignore patterns set
- [ ] `npm run lint` executes without config errors
- [ ] `npm run lint:fix` applies safe fixes only
- [ ] Changes committed to repository

## Reference

Main plan section: "Configuration Files → 4. Enhanced ESLint Configuration"

## Testing

1. Test linting:
   ```bash
   cd frontend
   npm run lint
   ```

2. Create test file with violations:
   ```bash
   echo 'const x = 1; console.log(x);' > test-lint.ts
   npm run lint test-lint.ts
   # Should show: quotes, semi, console warnings
   ```

3. Test auto-fix:
   ```bash
   npm run lint:fix test-lint.ts
   cat test-lint.ts
   # Should have single quotes and semicolon
   rm test-lint.ts
   ```

## Notes

- Some rules will trigger warnings on existing code (that's expected)
- Auto-fixable rules will be used in pre-commit hooks (Phase 3)
- Non-auto-fixable rules (like `no-explicit-any`) are warnings to guide fixes
- This config integrates with CodeRabbit's ESLint tool integration
