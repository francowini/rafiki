# Task 4.4: Implement Tier Categorization

**Phase**: 4 - Claude Code Command
**Estimated Time**: 3 hours
**Priority**: Critical
**Dependencies**: Task 4.3 (comment fetching implemented)

## Objective

Implement the logic to categorize CodeRabbit suggestions into three tiers: Tier 1 (safe auto-fix), Tier 2 (requires approval), and Tier 3 (manual review only). This is the core intelligence of the `/coderabbit-review` command.

## Steps

1. **Define Tier 3 protected paths**

   Create a list of paths that should NEVER be auto-fixed:

   **Backend:**
   - `business/types/**` - Business validation logic
   - `business/domain/**/model.go` - API contracts
   - `app/sdk/auth/**` - Authentication code
   - `foundation/keystore/**` - Cryptographic keys

   **Frontend:**
   - `lib/auth-context.tsx` - Auth state
   - `lib/api.ts` - API client

   **Keywords indicating Tier 3:**
   - "security", "sql", "jwt", "password", "crypto", "token"

2. **Define Tier 1 indicators**

   Comments that indicate safe auto-fix:
   - Contains: "formatting", "import", "whitespace", "indentation", "prettier", "eslint", "gofmt"
   - File NOT in protected paths
   - No function signature changes
   - No logic changes

3. **Define Tier 2 indicators**

   Everything else that's not Tier 1 or Tier 3:
   - Contains: "error handling", "unused", "simplify", "refactor"
   - Changes to exported functions
   - Logic improvements
   - Type safety improvements

4. **Implement categorization function**

   Create JavaScript/bash logic that:
   ```javascript
   function categorizeSuggestion(comment) {
     const filePath = comment.path;
     const body = comment.body.toLowerCase();

     // Tier 3 (highest priority - manual only)
     if (isProtectedPath(filePath) || hasSecurityKeywords(body)) {
       return { tier: 3, reason: 'Security-sensitive or business logic' };
     }

     // Tier 2 (requires approval)
     if (hasLogicChangeIndicators(body)) {
       return { tier: 2, reason: 'Logic change requires review' };
     }

     // Tier 1 (safe auto-fix)
     if (hasFormattingIndicators(body)) {
       return { tier: 1, reason: 'Safe formatting/style fix' };
     }

     // Conservative default
     return { tier: 2, reason: 'Uncertain, requires review' };
   }
   ```

5. **Implement helper functions**

   **isProtectedPath(filePath)**:
   ```javascript
   function isProtectedPath(filePath) {
     const protectedPatterns = [
       /business\/types\//,
       /business\/domain\/.*\/model\.go$/,
       /app\/sdk\/auth\//,
       /foundation\/keystore\//,
       /lib\/auth-context\.tsx$/,
       /lib\/api\.ts$/,
     ];
     return protectedPatterns.some(pattern => pattern.test(filePath));
   }
   ```

   **hasSecurityKeywords(body)**:
   ```javascript
   function hasSecurityKeywords(body) {
     const keywords = ['security', 'sql', 'jwt', 'password', 'crypto', 'token', 'auth'];
     return keywords.some(keyword => body.includes(keyword));
   }
   ```

   **hasLogicChangeIndicators(body)**:
   ```javascript
   function hasLogicChangeIndicators(body) {
     const indicators = [
       'error handling',
       'unused function',
       'simplify logic',
       'refactor',
       'remove unused',
     ];
     return indicators.some(indicator => body.includes(indicator));
   }
   ```

   **hasFormattingIndicators(body)**:
   ```javascript
   function hasFormattingIndicators(body) {
     const indicators = [
       'formatting',
       'import',
       'whitespace',
       'indentation',
       'prettier',
       'eslint',
       'gofmt',
     ];
     return indicators.some(indicator => body.includes(indicator));
   }
   ```

6. **Process all comments through categorization**
   ```bash
   # In the slash command
   COMMENTS=$(gh api repos/{owner}/{repo}/pulls/$PR_NUMBER/comments)

   # Categorize each comment
   for comment in $COMMENTS; do
     TIER=$(categorizeSuggestion "$comment")
     # Store in tier-specific arrays
     if [ "$TIER" = "1" ]; then
       TIER1_COMMENTS+=("$comment")
     elif [ "$TIER" = "2" ]; then
       TIER2_COMMENTS+=("$comment")
     else
       TIER3_COMMENTS+=("$comment")
     fi
   done
   ```

7. **Add logging and debugging**
   ```bash
   echo "Categorization results:"
   echo "- Tier 1 (auto-fix): ${#TIER1_COMMENTS[@]} issues"
   echo "- Tier 2 (review): ${#TIER2_COMMENTS[@]} issues"
   echo "- Tier 3 (manual): ${#TIER3_COMMENTS[@]} issues"
   ```

8. **Test categorization with sample comments**

   Create test cases:
   - Test comment about formatting → Should be Tier 1
   - Test comment about error wrapping → Should be Tier 2
   - Test comment in `business/types/` → Should be Tier 3
   - Test comment with "security" keyword → Should be Tier 3

9. **Add categorization reasoning**

   Store why each comment was categorized:
   ```javascript
   return {
     tier: 2,
     reason: 'Logic change requires review',
     comment: comment,
     indicators: ['error handling keyword found'],
   };
   ```

10. **Commit the categorization logic**
    ```bash
    git add .claude/commands/coderabbit-review.md
    git commit -m "feat: implement tier categorization for CodeRabbit comments

    Add smart categorization logic:
    - Tier 1: Safe formatting/style fixes (auto-apply)
    - Tier 2: Logic changes requiring approval (interactive)
    - Tier 3: Security/business logic (manual only)

    Protected paths:
    - business/types/** (validation logic)
    - app/sdk/auth/** (security)
    - lib/auth-context.tsx (frontend auth)

    Conservative default: When uncertain, use Tier 2.
    Part of CodeRabbit automation (Phase 4)."
    ```

## Expected Output

Categorization logic that:
- Accurately identifies Tier 1 (formatting) issues
- Flags Tier 3 (security/business) issues
- Defaults to Tier 2 for uncertain cases
- Provides clear reasoning for each categorization

## Success Criteria

- [ ] Categorization function implemented
- [ ] Protected path detection working
- [ ] Security keyword detection working
- [ ] Formatting indicator detection working
- [ ] Logic change indicator detection working
- [ ] All comments categorized into correct tiers
- [ ] Categorization reasoning provided
- [ ] Conservative default (Tier 2) for uncertain cases
- [ ] Test cases pass for all three tiers
- [ ] Code committed to command file

## Reference

Main plan section: "Claude Code Slash Command Design → Phase 3: Categorization Logic"

## Testing Checklist

Test with sample comments:
- [ ] "Fix indentation" → Tier 1
- [ ] "Organize imports" → Tier 1
- [ ] "Use %w for error wrapping" → Tier 2
- [ ] "Remove unused function" → Tier 2
- [ ] Change in `business/types/intensity.go` → Tier 3
- [ ] "Fix JWT token validation" → Tier 3 (security keyword)
- [ ] Change in `lib/api.ts` → Tier 3 (protected path)

## Sample Test

```javascript
// Test categorization
const testComments = [
  { path: 'main.go', body: 'Fix indentation and formatting' },
  { path: 'handler.go', body: 'Use %w for error wrapping instead of %s' },
  { path: 'business/types/intensity/intensity.go', body: 'Simplify validation' },
  { path: 'app/sdk/auth/auth.go', body: 'Improve JWT token handling' },
];

testComments.forEach(comment => {
  const result = categorizeSuggestion(comment);
  console.log(`${comment.path}: Tier ${result.tier} - ${result.reason}`);
});

// Expected output:
// main.go: Tier 1 - Safe formatting/style fix
// handler.go: Tier 2 - Logic change requires review
// business/types/intensity/intensity.go: Tier 3 - Security-sensitive or business logic
// app/sdk/auth/auth.go: Tier 3 - Security-sensitive or business logic
```

## Notes

**Conservative Approach**:
- When in doubt, categorize as Tier 2 (requires approval)
- Never accidentally auto-fix security-sensitive code
- Better to ask for approval than break something

**Tuning**:
- Monitor false positives/negatives in first week
- Adjust keywords and patterns based on real usage
- Add more protected paths as needed

**Edge Cases**:
- Multiple indicators (e.g., formatting + logic) → Use higher tier
- Ambiguous comments → Default to Tier 2
- Unknown file paths → Default to Tier 2
