# Task 1.2: Create Go Linter Configuration

**Phase**: 1 - Foundation Setup
**Estimated Time**: 1 hour
**Priority**: High
**Dependencies**: None

## Objective

Create `.golangci.yml` configuration file to define which Go linters are enabled, which can auto-fix (Tier 1), and specific settings for the Rafiki codebase.

## Steps

1. **Create the configuration file**
   ```bash
   touch .golangci.yml
   ```

2. **Configure run settings**
   - Set timeout to 10 minutes
   - Enable test file linting
   - Skip vendor directory
   - Use readonly modules download mode

3. **Configure auto-fixable linters (Tier 1)**
   Enable these linters that support `--fix`:
   - `gofmt` - Go formatting
   - `gofumpt` - Stricter formatting
   - `goimports` - Import organization
   - `gci` - Import grouping
   - `gosimple` - Simplification suggestions
   - `ineffassign` - Ineffectual assignments
   - `whitespace` - Whitespace issues
   - `canonicalheader` - HTTP header formatting

4. **Configure important linters (no auto-fix)**
   Enable for safety checks:
   - `errcheck` - Unchecked errors
   - `govet` - Go vet analysis
   - `staticcheck` - Static analysis
   - `typecheck` - Type checking
   - `bodyclose` - HTTP body close checks
   - `exportloopref` - Loop variable captures
   - `gocritic` - Various checks
   - `revive` - Golint replacement
   - `misspell` - Spelling
   - `unconvert` - Unnecessary conversions
   - `unparam` - Unused parameters
   - `nakedret` - Naked returns

5. **Configure linter-specific settings**
   - `gofumpt`: Set module path to `github.com/francowini/rafiki`
   - `goimports`: Set local-prefixes to `github.com/francowini/rafiki`
   - `gci`: Configure sections (standard, default, prefix)
   - `errcheck`: Check type assertions and blank assignments
   - `depguard`: Deny deprecated `io/ioutil` package

6. **Configure issue exclusions**
   - Exclude `errcheck`, `gosec`, `dupl` from test files
   - Exclude `unused` checks from `business/types/` (public API)
   - Exclude all linters from `vendor/`

7. **Disable overly strict linters**
   - `exhaustruct` - Too strict for our use case
   - `varnamelen` - Conflicts with idiomatic Go
   - `gochecknoglobals` - We use globals appropriately
   - `gochecknoinits` - We use init() appropriately

## Expected Output

File: `.golangci.yml` at repository root

Key sections:
```yaml
run:
  timeout: 10m
linters:
  enable: [list of auto-fixable and important linters]
  disable: [list of too-strict linters]
linters-settings:
  gofumpt:
    module-path: github.com/francowini/rafiki
  goimports:
    local-prefixes: github.com/francowini/rafiki
```

## Success Criteria

- [ ] `.golangci.yml` file created in repository root
- [ ] All Tier 1 auto-fixable linters enabled
- [ ] Important safety linters enabled
- [ ] Module path correctly set to `github.com/francowini/rafiki`
- [ ] Import organization configured properly
- [ ] Appropriate exclusions for test files and business types
- [ ] File committed to repository

## Reference

See full configuration template in main plan: Section "Configuration Files → 2. Go Linter Configuration"

## Testing

1. Run golangci-lint locally:
   ```bash
   golangci-lint run
   ```

2. Test auto-fix on Tier 1 linters:
   ```bash
   golangci-lint run --fix --enable-only=gofmt,goimports,gci,gosimple,ineffassign,whitespace
   ```

3. Verify no errors on existing codebase (or document expected issues)

## Notes

- This configuration separates auto-fixable (Tier 1) from manual review linters
- The Tier 1 linters will be used in pre-commit hooks and CI auto-fix
- Business types directory is excluded from unused checks as it's a public API
