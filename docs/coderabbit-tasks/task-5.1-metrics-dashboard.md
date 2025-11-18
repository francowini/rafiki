# Task 5.1: Setup Metrics Dashboard

**Phase**: 5 - Monitoring & Optimization
**Estimated Time**: Ongoing (2 hours/week)
**Priority**: Medium
**Dependencies**: Phases 1-4 complete (automation running)

## Objective

Create a simple metrics tracking system to monitor the effectiveness of CodeRabbit automation, identify areas for improvement, and demonstrate value to the team.

## Steps

1. **Create metrics tracking document**
   ```bash
   mkdir -p docs/metrics
   touch docs/metrics/weekly-dashboard.md
   ```

2. **Define key metrics to track**

   **Automation Metrics:**
   - PRs opened this week
   - Tier 1 auto-fixes applied (count)
   - Tier 2 suggestions reviewed (count)
   - Tier 2 accepted rate (%)
   - Tier 3 manual reviews (count)

   **Performance Metrics:**
   - Average PR review time (hours)
   - Average time to merge (days)
   - CI/CD duration (minutes)

   **Quality Metrics:**
   - CodeRabbit suggestions per PR (trend)
   - False positive categorizations (count)
   - Pre-commit hook bypasses (count)
   - CI failures due to auto-fix (count)

3. **Create weekly dashboard template**

   ```markdown
   # Week of YYYY-MM-DD

   ## Automation
   - PRs opened: X
   - Tier 1 auto-fixes applied: X/Y PRs (Z%)
   - Tier 2 suggestions reviewed: X
   - Tier 2 accepted: X/Y (Z%)
   - Tier 3 manual reviews: X

   ## Performance
   - Average PR review time: X.X hours (↑↓ from X.X)
   - Average time to merge: X.X days (↑↓ from X.X)
   - CI/CD duration: X.X minutes ✅/❌

   ## Issues
   - False positive categorizations: X
   - Pre-commit hook bypasses: X
   - CI failures due to auto-fix: X

   ## Highlights
   - [Notable achievements or issues this week]

   ## Action Items
   - [ ] [Things to improve next week]
   ```

4. **Collect baseline metrics**

   Before automation (for comparison):
   ```bash
   # Get PR metrics from last month (pre-automation)
   gh pr list --state merged --limit 20 --json number,createdAt,mergedAt

   # Calculate average time to merge
   # Document baseline metrics
   ```

5. **Set up data collection methods**

   **Manual Collection (Weekly):**
   - Count PRs: `gh pr list --state all --search "created:YYYY-MM-DD..YYYY-MM-DD"`
   - Review GitHub Actions logs for auto-fix commits
   - Check CodeRabbit comments for suggestion counts

   **Automated Queries:**
   ```bash
   # Script: scripts/collect-metrics.sh

   # Get weekly PR count
   gh pr list --state all --search "created:>=$(date -d '7 days ago' +%Y-%m-%d)" --json number | jq length

   # Get auto-fix commits
   gh pr list --state all --search "created:>=$(date -d '7 days ago' +%Y-%m-%d)" \
     --json number,commits | jq '[.[] | .commits[] | select(.messageHeadline | contains("auto-fix"))] | length'
   ```

6. **Create weekly review process**

   Every Monday morning:
   1. Run metrics collection script
   2. Update weekly dashboard document
   3. Review trends and anomalies
   4. Identify action items for the week
   5. Share summary with team

7. **Track categorization accuracy**

   Create a log for misclassifications:
   ```markdown
   # Categorization Issues Log

   ## YYYY-MM-DD
   - **False Positive Tier 1**: [PR#] [File] - [Why it shouldn't be Tier 1]
   - **False Negative Tier 3**: [PR#] [File] - [Why it should be Tier 3]
   - **Action**: [How to improve categorization]
   ```

8. **Set up alerts for anomalies**

   Watch for:
   - CI failure rate >5%
   - Hook bypass rate >20%
   - Tier 2 rejection rate >30%
   - Average review time increasing

9. **Create monthly summary**

   At end of each month:
   - Aggregate weekly dashboards
   - Calculate month-over-month trends
   - Report savings (time, effort)
   - Present to team

10. **Share insights with team**

    Include in stand-ups or team meetings:
    - "This week: 8 PRs, 6 had auto-fixes, saved ~2 hours of manual formatting"
    - "Tier 2 acceptance rate 85% - categorization working well"
    - "Average PR review time down from 3.2 hours to 2.1 hours"

## Expected Output

### Weekly Dashboard Document
Regular updates showing automation effectiveness

### Metrics Collection Script
Automated queries to reduce manual effort

### Insights and Actions
Data-driven improvements to the system

## Success Criteria

- [ ] Weekly dashboard template created
- [ ] Baseline metrics collected (pre-automation)
- [ ] First week's metrics documented
- [ ] Metrics collection script created
- [ ] Weekly review process established
- [ ] Team receives regular updates
- [ ] Action items tracked and completed
- [ ] Month-over-month trends visible

## Reference

Main plan section: "Success Metrics"

## Metrics to Emphasize

**Time Savings:**
- Hours saved on formatting per week
- Reduction in PR review time
- Faster time to merge

**Quality Improvements:**
- Reduction in formatting issues
- Consistent code style
- Fewer linting errors

**Automation Success:**
- High Tier 1 auto-fix rate
- High Tier 2 acceptance rate
- Low false positive rate

## Sample Weekly Dashboard

```markdown
# Week of 2025-11-18

## Automation
- PRs opened: 12
- Tier 1 auto-fixes applied: 8/12 PRs (67%)
- Tier 2 suggestions reviewed: 23
- Tier 2 accepted: 18/23 (78%)
- Tier 3 manual reviews: 5

## Performance
- Average PR review time: 2.1 hours (↓ from 2.8 hours)
- Average time to merge: 1.3 days (↓ from 1.7 days)
- CI/CD duration: 4.2 minutes ✅ (target: <5 min)

## Issues
- False positive categorizations: 2
  - PR#42: business/domain/momentbus.go flagged as Tier 2, should be Tier 3
  - Action: Add momentbus path to protected list
- Pre-commit hook bypasses: 3 (WIP commits, acceptable)
- CI failures due to auto-fix: 0 ✅

## Highlights
- First full week with all automation active
- Team feedback positive (8/10 satisfaction)
- Saved ~10 hours of manual formatting this week

## Action Items
- [x] Add momentbus to Tier 3 protected paths
- [ ] Create quick reference guide for categorization
- [ ] Schedule team retrospective for next week
```

## Notes

- Keep metrics simple and actionable
- Focus on trends, not absolute numbers
- Celebrate wins (time saved, quality improved)
- Be transparent about issues and fixes
- Use data to justify continued investment in automation
