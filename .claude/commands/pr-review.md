# PR Review - Issue Correction & Documentation

Review the current branch diff against main, receive user-reported issues, fix them, and document important lessons learned.

**Usage**: `/pr-review`

## Overview

This command facilitates a collaborative PR review process where:
1. Claude shows the diff of the current branch vs main
2. User raises issues they've found
3. Claude corrects each issue
4. Issues of severity Medium, Major, or Critical are added to the appropriate errors-to-avoid documentation

## Execution Flow

### Step 1: Show PR Diff

Run the following to get the diff:
```bash
git diff main...HEAD --stat
git diff main...HEAD
```

Present a summary of:
- Files changed (with line counts)
- Key changes in each area (backend, frontend, devops)

### Step 2: Receive Issues

Ask the user to describe issues they've found. For each issue, gather:
- **Description**: What is the problem?
- **Location**: Which file(s) and line(s)?
- **Severity**: Low, Medium, Major, or Critical

### Step 3: Fix Each Issue

For each issue reported:
1. Read the affected file(s)
2. Understand the problem
3. Apply the fix using Edit tool
4. Verify the fix compiles/builds
5. Confirm with user

### Step 4: Document Lessons Learned

**CRITICAL**: For issues with severity **Medium, Major, or Critical**, add them to the appropriate errors-to-avoid file:

- **Backend issues (Go)** → `devs/errors-to-avoid-backend.md`
- **Frontend issues (TypeScript/React)** → `devs/errors-to-avoid-frontend.md`

**Documentation Format:**

```markdown
## [Number]. [Category]: [Short Title]

### Severity: [Medium | Major | Critical] ([Impact Description])

### Problem

[Describe what went wrong and why it's problematic]

### Bad Example

```[language]
// BAD: [explanation]
[code that caused the issue]
```

### Good Example

```[language]
// GOOD: [explanation]
[corrected code]
```

### Checklist

- [ ] [Actionable item to prevent this in the future]
- [ ] [Another actionable item if needed]
```

### Step 5: Commit Fixes

After all issues are fixed:
1. Stage the changes
2. Commit with message: `fix: address PR review feedback`
3. Push to the branch

## Severity Guidelines

| Severity | Description | Document? |
|----------|-------------|-----------|
| **Low** | Code style, minor improvements | No |
| **Medium** | Bug potential, maintainability issue | Yes |
| **Major** | Actual bug, security concern, data integrity | Yes |
| **Critical** | Security vulnerability, data loss, crash | Yes |

## Issue Categories

Common categories for the errors-to-avoid docs:

**Backend:**
- Security (ownership validation, injection)
- Error Handling (sentinel errors, proper status codes)
- Data Validation (UTF-8, bounds checking)
- SQL (query building, performance)
- Thread Safety (race conditions)
- API Design (status codes, contracts)

**Frontend:**
- Security (XSS, injection)
- Async (race conditions, stale data)
- State Management (sync issues)
- Accessibility (ARIA, keyboard)
- Data Handling (validation, parsing)
- HTML Semantics (nesting rules)

## Example Session

```
User: /pr-review

Claude: [Shows diff summary]
        Files changed: 26 files (+1351, -149)

        Backend: entitystatus type, migrations, valuebus/lifevisionbus updates
        Frontend: ReassignmentDialog, ValueCard, ValueList updates

        What issues have you found?

User: The AlertDialogDescription has nested <p> tags which causes hydration errors

Claude: Issue #1: HTML Semantics - Nested <p> tags in AlertDialogDescription
        Severity: Medium (Hydration Error)

        [Applies fix]

        This is a Medium severity issue. Adding to devs/errors-to-avoid-frontend.md...

        Any other issues?

User: The API returns 400 but frontend checks for 409

Claude: Issue #2: API Design - Inconsistent HTTP Status Codes
        Severity: Major (Integration Bug)

        [Applies fix to backend to return 409]

        This is a Major severity issue. Adding to devs/errors-to-avoid-backend.md...
```

## Output

After all issues are addressed:
1. Summary of fixes applied
2. List of issues added to errors-to-avoid docs
3. Commit and push confirmation

Execute this PR review workflow interactively with the user.
