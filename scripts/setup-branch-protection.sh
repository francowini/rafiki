#!/bin/bash
#
# GitHub Branch Protection Setup Script
#
# This script configures branch protection rules for the main branch via GitHub API.
#
# Prerequisites:
#   - GitHub CLI (gh) installed and authenticated
#   - Repository: francowini/rafiki
#   - User must have admin access to the repository
#
# Usage:
#   ./scripts/setup-branch-protection.sh
#

set -e

REPO="francowini/rafiki"
BRANCH="main"

echo "Setting up branch protection for $REPO:$BRANCH"
echo "================================================"
echo ""

# Check if gh CLI is installed and authenticated
if ! command -v gh &> /dev/null; then
    echo "❌ Error: GitHub CLI (gh) is not installed."
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ Error: Not authenticated with GitHub CLI."
    echo "Run: gh auth login"
    exit 1
fi

echo "✅ GitHub CLI authenticated"
echo ""

# Create branch protection rule
echo "Configuring branch protection rules..."
echo ""

# Note: As of Phase 1, status checks are not yet created (they will be in Phase 2)
# This script sets up the foundation and can be updated in Phase 2

cat << 'EOF' > /tmp/branch-protection.json
{
  "required_status_checks": null,
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1,
    "require_last_push_approval": false
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": true
}
EOF

# Apply the branch protection
gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "/repos/$REPO/branches/$BRANCH/protection" \
  --input /tmp/branch-protection.json

rm /tmp/branch-protection.json

echo ""
echo "✅ Branch protection configured successfully!"
echo ""
echo "Configuration applied:"
echo "  - Required pull request reviews: 1 approval"
echo "  - Dismiss stale reviews: enabled"
echo "  - Required linear history: enabled"
echo "  - Required conversation resolution: enabled"
echo "  - Force pushes: disabled"
echo "  - Branch deletion: disabled"
echo ""
echo "⚠️  Note: Status checks will be added in Phase 2 (after GitHub Actions are set up)"
echo ""
echo "You can view the settings at:"
echo "https://github.com/$REPO/settings/branches"
