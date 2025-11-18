# Task 1.3: Create Prettier Configuration

**Phase**: 1 - Foundation Setup
**Estimated Time**: 1 hour
**Priority**: High
**Dependencies**: None

## Objective

Create Prettier configuration files (`.prettierrc` and `.prettierignore`) in the frontend directory to establish consistent code formatting for TypeScript, TSX, JSON, and Markdown files.

## Steps

1. **Create Prettier config file**
   ```bash
   touch frontend/.prettierrc
   ```

2. **Configure formatting rules**
   Add JSON configuration with:
   - Semicolons: Always use
   - Trailing commas: ES5 compatible
   - Quotes: Single quotes for JS/TS
   - Print width: 100 characters
   - Tab width: 2 spaces
   - Use spaces (not tabs)
   - Arrow parens: Always
   - Line endings: LF (Unix-style)
   - Bracket spacing: Enabled
   - JSX quotes: Double quotes
   - JSX bracket same line: False

3. **Create Prettier ignore file**
   ```bash
   touch frontend/.prettierignore
   ```

4. **Configure ignore patterns**
   Exclude from formatting:
   - Dependencies: `node_modules`, `.pnp`, `.pnp.js`
   - Build output: `.next`, `out`, `build`, `dist`
   - Cache: `.turbo`, `.cache`
   - Environment: `.env*` files
   - Lock files: `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`
   - Generated/minified: `*.min.js`, `*.min.css`

5. **Install Prettier as dev dependency**
   ```bash
   cd frontend
   npm install --save-dev prettier
   ```

## Expected Output

### File 1: `frontend/.prettierrc`
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

### File 2: `frontend/.prettierignore`
Contains ignore patterns for dependencies, build output, cache, etc.

## Success Criteria

- [ ] `frontend/.prettierrc` created with proper JSON formatting rules
- [ ] `frontend/.prettierignore` created with comprehensive ignore patterns
- [ ] Prettier installed as dev dependency in `frontend/package.json`
- [ ] Config uses single quotes for consistency
- [ ] 100-character line width set (readable on modern monitors)
- [ ] LF line endings enforced (Unix/Linux compatibility)
- [ ] Files committed to repository

## Reference

See full configuration templates in main plan: Section "Configuration Files → 3. Prettier Configuration"

## Testing

1. Test Prettier on a sample file:
   ```bash
   cd frontend
   npx prettier --check "app/**/*.tsx"
   ```

2. Check that ignored paths are properly excluded:
   ```bash
   npx prettier --check "."
   # Should not check node_modules, .next, etc.
   ```

3. Verify formatting rules:
   ```bash
   echo "const x={a:1,b:2}" > test.ts
   npx prettier --write test.ts
   cat test.ts
   # Should have spaces, single quotes, semicolon
   rm test.ts
   ```

## Notes

- Don't run `prettier --write` on entire codebase yet (that's Task 1.4)
- This task only creates the configuration files
- Prettier will be run via npm scripts (added in Task 1.5)
- This config matches common Next.js/TypeScript conventions
