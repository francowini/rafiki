# Task 1.1: Create CodeRabbit Configuration

**Phase**: 1 - Foundation Setup
**Estimated Time**: 2 hours
**Priority**: High
**Dependencies**: None

## Objective

Create the `.coderabbit.yaml` configuration file to customize CodeRabbit's review behavior for the Rafiki project, including path-specific instructions for business logic and security-sensitive code.

## Steps

1. **Create the configuration file**
   ```bash
   touch .coderabbit.yaml
   ```

2. **Add basic configuration**
   - Set language to en-US
   - Enable assertive review profile
   - Enable auto-review for PRs to main branch
   - Disable reviews for draft PRs

3. **Configure path filters**
   Exclude from review:
   - `vendor/**`
   - `frontend/node_modules/**`
   - `frontend/.next/**`
   - `**/*.pb.go` (generated protobuf files)
   - `**/*_generated.go`

4. **Configure tool integrations**
   - Enable golangci-lint with config file `.golangci.yml`
   - Enable ESLint with config file `frontend/eslint.config.mjs`

5. **Add path-specific instructions**
   Critical paths that require special handling:
   - `business/types/**` - Business validation types (never auto-fix)
   - `business/domain/**/model.go` - API contract models
   - `app/sdk/auth/**` - Authentication code (security-sensitive)
   - `foundation/keystore/**` - Cryptographic keys (security-sensitive)
   - `frontend/lib/auth-context.tsx` - Client auth state
   - `frontend/lib/api.ts` - API client
   - `api/services/partners/mux/**` - HTTP routing patterns
   - `frontend/app/**` - Next.js 16 App Router conventions

6. **Configure labels**
   - Add "coderabbit" label to reviewed PRs

## Expected Output

File: `.coderabbit.yaml` at repository root

Key sections:
```yaml
language: "en-US"
reviews:
  profile: "assertive"
  auto_review:
    enabled: true
  path_filters: [...]
  tools: [...]
  path_instructions: [...]
```

## Success Criteria

- [ ] `.coderabbit.yaml` file created in repository root
- [ ] All path filters properly exclude generated/dependency code
- [ ] Tool integrations configured for golangci-lint and ESLint
- [ ] Path-specific instructions added for all security-sensitive directories
- [ ] Instructions match Rafiki's Go 1.22+ and Next.js 16 patterns
- [ ] File committed to repository

## Reference

See full configuration template in main plan: Section "Configuration Files → 1. CodeRabbit Configuration"

## Testing

1. Create a test PR with changes to various paths
2. Verify CodeRabbit reviews the PR
3. Check that path-specific instructions appear in comments
4. Verify excluded paths are not reviewed

## Notes

- The assertive profile makes CodeRabbit more proactive with suggestions
- Path instructions help CodeRabbit understand project-specific patterns
- These instructions will inform the Tier categorization in Phase 4
