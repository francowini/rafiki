# Task 1.7: Setup Branch Protection

**Phase**: 1 - Foundation Setup
**Estimated Time**: 1.5 hours
**Priority**: High
**Dependencies**: Tasks 1.1-1.6 (configs must exist first)

## Objective

Configure GitHub branch protection rules for the `main` branch to enforce code quality gates, require PR reviews, and prevent direct pushes. Create an automation script for reproducible setup.

## Steps

### Part 1: Create Automation Script (1 hour)

1. **Create scripts directory if needed**
   ```bash
   mkdir -p scripts
   ```

2. **Create branch protection script**
   ```bash
   touch scripts/setup-branch-protection.sh
   chmod +x scripts/setup-branch-protection.sh
   ```

3. **Add script content**
   The script should:
   - Check for `GITHUB_TOKEN` environment variable
   - Use GitHub API to configure branch protection for `main`
   - Set required status checks (to be added in Phase 2)
   - Require PR reviews (1 approval)
   - Enforce for administrators
   - Require linear history
   - Prevent force pushes and deletions

4. **Add usage documentation to script**
   Include comments explaining:
   - How to create a GitHub token
   - Required token scopes (repo, admin:repo_hook)
   - How to run the script

### Part 2: Configure via GitHub UI (30 minutes)

Alternatively, configure manually via GitHub web interface:

1. **Navigate to repository settings**
   - Go to: https://github.com/francowini/rafiki/settings/branches

2. **Add branch protection rule**
   - Branch name pattern: `main`

3. **Configure protection settings**

   ✅ **Require a pull request before merging**
   - Required approvals: 1
   - Dismiss stale reviews: ✅
   - Require review from Code Owners: ☐ (optional)

   ✅ **Require status checks to pass before merging**
   - Require branches to be up to date: ✅
   - Status checks (will add in Phase 2):
     - `lint-backend`
     - `test-backend`
     - `lint-frontend`
     - `typecheck-frontend`
     - `test-frontend`

   ✅ **Require conversation resolution before merging**

   ✅ **Require linear history**

   ✅ **Do not allow bypassing the above settings**

   ✅ **Restrict who can push to matching branches**
   - Leave empty (blocks all direct pushes)

   ☐ **Allow force pushes**: NEVER

   ☐ **Allow deletions**: NEVER

4. **Save the protection rule**

### Part 3: Testing and Documentation

1. **Test branch protection**
   ```bash
   # Try to push directly to main (should fail)
   git checkout main
   git commit --allow-empty -m "test: branch protection"
   git push origin main
   # Expected: Error - branch protection active

   # Cleanup
   git reset HEAD~1
   ```

2. **Document the setup**
   Add to project documentation (README or CONTRIBUTING.md):
   - Branch protection is active on `main`
   - All changes must go through PRs
   - PRs require 1 approval
   - Status checks will be enforced (Phase 2)

3. **Commit the script**
   ```bash
   git add scripts/setup-branch-protection.sh
   git commit -m "chore: add branch protection setup script

   Add automation script for configuring GitHub branch protection:
   - Requires PR reviews (1 approval)
   - Requires status checks (configured in Phase 2)
   - Enforces linear history
   - Prevents direct pushes to main
   - Prevents force pushes and deletions

   Can be run via script or configured manually via GitHub UI.
   Part of CodeRabbit automation setup (Phase 1)."

   # Push via PR since main is now protected
   git checkout -b chore/branch-protection-script
   git push origin chore/branch-protection-script
   gh pr create --title "Add branch protection script" --body "See commit message"
   ```

## Expected Output

### Script: `scripts/setup-branch-protection.sh`
Executable script that configures branch protection via GitHub API.

### Branch Protection Active
- Direct pushes to `main` blocked
- PRs require 1 approval
- Linear history enforced
- Force pushes disabled

## Success Criteria

- [ ] Branch protection script created and executable
- [ ] Script has clear documentation/comments
- [ ] Branch protection active on `main` branch
- [ ] Direct pushes to `main` blocked (tested)
- [ ] PR reviews required (1 approval)
- [ ] Linear history enforced
- [ ] Force pushes and deletions disabled
- [ ] Script committed to repository
- [ ] Team notified of new workflow (PRs required)

## Reference

Main plan sections:
- "GitHub Branch Protection → Configuration via API Script"
- "GitHub Branch Protection → Configuration via GitHub UI"

## Testing Checklist

- [ ] Direct push to main fails with protection error
- [ ] Can create PR targeting main
- [ ] PR cannot be merged without approval
- [ ] PR cannot be merged with unresolved conversations
- [ ] Merge creates linear history (no merge commits)

## Notes

**IMPORTANT**:
- Status checks will be added in Phase 2 (after workflows exist)
- For now, only PR review requirement is enforced
- Once status checks are added (Phase 2), update protection to require them

**GitHub Token Scopes**:
If using script, token needs:
- `repo` - Full repository access
- `admin:repo_hook` - Repository webhook access

**Alternative Approach**:
If you prefer UI configuration over script, that's fine. The script is for:
- Reproducibility (can reapply if settings change)
- Documentation (script shows exactly what's configured)
- Automation (can be run in CI/setup scripts)
