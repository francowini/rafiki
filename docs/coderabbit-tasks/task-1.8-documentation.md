# Task 1.8: Update Documentation

**Phase**: 1 - Foundation Setup
**Estimated Time**: 2 hours
**Priority**: Medium
**Dependencies**: Tasks 1.1-1.7 (all Phase 1 work complete)

## Objective

Update project documentation (README.md and create CONTRIBUTING.md) to reflect the new CodeRabbit automation workflow, branch protection rules, and code quality standards.

## Steps

### Part 1: Update README.md (45 minutes)

1. **Add Code Quality section**
   Add a new section after "Architecture" or before "Deployment":

   ```markdown
   ## Code Quality

   This project uses automated code quality tools:

   - **CodeRabbit Pro**: AI-powered code review on all PRs
   - **Pre-commit hooks**: Auto-formatting before commits (Phase 3)
   - **GitHub Actions**: Automated linting and testing on PRs
   - **Branch protection**: Main branch requires PR reviews and passing checks

   ### Quick Start

   ```bash
   # Install developer tools (one-time setup)
   ./scripts/setup-tools.sh

   # Install git hooks (one-time setup)
   ./scripts/install-git-hooks.sh

   # Your commits will now auto-format code
   git commit -m "feat: my feature"
   ```

   ### Formatting and Linting

   **Backend (Go):**
   ```bash
   # Format code
   gofmt -w -s .
   goimports -w -local github.com/francowini/rafiki .

   # Lint
   golangci-lint run
   ```

   **Frontend (TypeScript/Next.js):**
   ```bash
   cd frontend

   # Format code
   npm run format

   # Lint
   npm run lint

   # Type check
   npm run typecheck

   # Run all checks
   npm run check
   ```

   See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.
   ```

2. **Update Development section**
   Add note about branch protection:
   ```markdown
   ### Development Workflow

   **IMPORTANT**: The `main` branch is protected. All changes must go through pull requests.

   1. Create a feature branch: `git checkout -b feature/my-feature`
   2. Make your changes
   3. Commit (pre-commit hooks will auto-format)
   4. Push and create a PR: `gh pr create`
   5. CodeRabbit will review your PR automatically
   6. Address review comments (use `/coderabbit-review` command)
   7. Get 1 approval from a team member
   8. Merge when all checks pass ✅
   ```

3. **Add workflow badges**
   At the top of README, add status badges (will activate in Phase 2):
   ```markdown
   # Rafiki

   [![Lint Backend](https://github.com/francowini/rafiki/workflows/PR%20Checks/badge.svg)](https://github.com/francowini/rafiki/actions)
   [![Lint Frontend](https://github.com/francowini/rafiki/workflows/PR%20Checks/badge.svg)](https://github.com/francowini/rafiki/actions)
   ```

### Part 2: Create CONTRIBUTING.md (1 hour 15 minutes)

1. **Create the file**
   ```bash
   touch CONTRIBUTING.md
   ```

2. **Add comprehensive content**
   Structure:
   - Welcome and overview
   - Development setup
   - Code style guidelines
   - Git workflow
   - Pull request process
   - CodeRabbit integration
   - Testing requirements
   - Questions and support

   Key sections to include:

   **Code Style**:
   - Go: Follow standard Go conventions, use gofmt/goimports
   - TypeScript: Follow Prettier/ESLint rules
   - Business types: Always use strong types from `business/types/`

   **Git Workflow**:
   - Branch naming: `feature/`, `fix/`, `chore/`, `docs/`
   - Commit messages: Conventional commits format
   - Branch protection: Main requires PRs

   **Pull Request Process**:
   - Create PR with clear description
   - CodeRabbit will review automatically
   - Address review comments
   - Use `/coderabbit-review` command (Phase 4)
   - Get 1 approval
   - Ensure all checks pass

   **Security and Business Logic**:
   - Never auto-fix `business/types/**`
   - Manual review required for `app/sdk/auth/**`
   - API contract changes need coordination

3. **Add examples**
   Include examples of:
   - Good commit messages
   - PR descriptions
   - How to use business types
   - How to run tests

### Part 3: Commit and Review

1. **Commit documentation changes**
   ```bash
   git add README.md CONTRIBUTING.md
   git commit -m "docs: update for CodeRabbit automation workflow

   Add documentation for:
   - Code quality tools (CodeRabbit, linters, formatters)
   - Branch protection and PR workflow
   - Development setup instructions
   - Contributing guidelines

   Create CONTRIBUTING.md with:
   - Code style guidelines
   - Git workflow
   - PR process
   - Security-sensitive code patterns

   Part of CodeRabbit automation setup (Phase 1)."
   ```

2. **Create PR for review**
   ```bash
   git checkout -b docs/coderabbit-workflow
   git push origin docs/coderabbit-workflow
   gh pr create --title "Update documentation for CodeRabbit workflow" \
     --body "Updates README and adds CONTRIBUTING.md. See commit for details."
   ```

## Expected Output

### README.md Updates
- New "Code Quality" section
- Updated "Development Workflow" section
- Workflow status badges (placeholders for Phase 2)

### New CONTRIBUTING.md
Complete contributing guide with:
- Development setup
- Code style guidelines
- Git and PR workflow
- CodeRabbit usage
- Security considerations

## Success Criteria

- [ ] README.md has Code Quality section
- [ ] README.md has updated Development Workflow
- [ ] Workflow badges added (will activate in Phase 2)
- [ ] CONTRIBUTING.md created with comprehensive guidelines
- [ ] Code style rules documented for Go and TypeScript
- [ ] Branch protection workflow explained
- [ ] PR process clearly documented
- [ ] Security-sensitive code patterns identified
- [ ] Changes committed and PR created
- [ ] Team reviews and approves documentation

## Reference

Main plan section: "Dependencies and Coordination → Team Communication"

## Template: CONTRIBUTING.md Structure

```markdown
# Contributing to Rafiki

## Welcome

## Development Setup
- Prerequisites
- Installation
- Running locally

## Code Style
- Go conventions
- TypeScript conventions
- Business types pattern

## Git Workflow
- Branch naming
- Commit messages
- Branch protection

## Pull Request Process
- Creating PRs
- CodeRabbit review
- Addressing feedback
- Getting approval

## Security and Business Logic
- Protected paths
- Manual review requirements

## Testing
- Running tests
- Writing tests

## Questions?
- Where to ask
- Documentation
```

## Notes

- Documentation is living - update as workflows evolve
- CONTRIBUTING.md should be linked in PR templates (Phase 2)
- Keep examples realistic and relevant to Rafiki
- Use clear, beginner-friendly language
- Link to main CodeRabbit plan for detailed technical info
