# Multi-Mind - Habits Tracker Development Team

Execute a collaborative analysis using two specialized subagents for the Rafiki habits tracker project.

**Usage**: `/multi-mind [topic/task]`

## Project Context

This is a habits tracker application (Rafiki/Topifier) built in Go with PostgreSQL, deployed on Hetzner servers. The two-agent team consists of:
1. **Backend Engineer**: Go development, API design, database operations
2. **DevOps Engineer**: Deployment, infrastructure, Docker, server management

## Implementation

Execute this two-specialist analysis using the Task tool to create independent subagents:

### Phase 1: Specialist Assignment & Research

Launch parallel subagents using the Task tool:

**Agent 1: Backend Engineer**
- Prompt: "As a backend engineer specializing in Go, analyze [topic] focusing on:
  - Go implementation patterns and best practices
  - PostgreSQL database design and queries
  - API endpoint design and HTTP routing (using Go 1.22+ patterns)
  - Service architecture (following the partner-service pattern)
  - Configuration management with ardanlabs/conf
  - Structured logging with foundation/logger
  - Error handling and validation
  - Review relevant code in api/services/partner/, foundation/, and business/
  Use Read, Glob, and Grep tools to examine the existing codebase."

**Agent 2: DevOps Engineer**
- Prompt: "As a DevOps engineer, analyze [topic] focusing on:
  - Docker and docker-compose configuration
  - Hetzner server deployment strategies
  - Service deployment and orchestration
  - Environment variable management
  - Health checks and monitoring
  - CORS and network configuration
  - Review deployment scripts and documentation in devops/ folder
  Use Read tool to examine deployment scripts, Docker files, and configuration files."

Each subagent operates independently with access to Read, Glob, Grep, and analysis tools.

**Specialist Domains**:
- **Backend**: Go codebase, API logic, database schemas, business logic
- **DevOps**: Infrastructure, deployment scripts, containerization, server configuration

### Phase 2: Cross-Pollination Round

After receiving both specialist reports, launch a second round of subagents:

```
Backend Engineer Review:
- Review DevOps findings for deployment considerations
- Identify how infrastructure choices affect code design
- Ensure API configuration aligns with deployment strategy
- Flag potential issues with containerization or environment setup

DevOps Engineer Review:
- Review Backend findings for deployment implications
- Identify infrastructure requirements for new features
- Ensure deployment scripts support code changes
- Flag potential performance or scaling concerns
```

### Phase 3: Synthesis & Final Recommendations

After both rounds:
- Collect all subagent outputs
- Synthesize backend and devops perspectives
- Identify coordination points between code and infrastructure
- Provide actionable implementation plan
- Highlight dependencies between backend work and devops work

## Anti-Repetition Mechanisms

**Moderator Responsibilities**:
- Track what has been thoroughly covered vs. what needs deeper exploration
- Redirect specialists away from rehashing previous points
- Push for new angles, deeper analysis, or broader implications
- Ensure backend and devops perspectives remain distinct and valuable

**Specialist Guidelines**:
- Build on previous round insights rather than restating them
- Backend focuses on code quality, maintainability, and Go best practices
- DevOps focuses on deployability, reliability, and operational excellence
- Cross-pollinate by considering the other specialist's domain concerns

## Output Protocol

```
=== RAFIKI HABITS TRACKER ANALYSIS: [Topic] ===
Specialists: Backend Engineer + DevOps Engineer

--- ROUND 1: INITIAL ANALYSIS ---
🔧 BACKEND ENGINEER ANALYSIS
[Go code analysis, API design, database considerations]

🚀 DEVOPS ENGINEER ANALYSIS
[Infrastructure analysis, deployment strategy, operational concerns]

--- ROUND 2: CROSS-POLLINATION ---
🔧 BACKEND ENGINEER REVIEW
[Review of DevOps findings and their impact on code]

🚀 DEVOPS ENGINEER REVIEW
[Review of Backend findings and their deployment implications]

--- FINAL SYNTHESIS ---
🎯 IMPLEMENTATION PLAN
[Step-by-step plan coordinating backend and devops work]

💻 BACKEND TASKS
[Specific Go code changes, database migrations, API updates]

🛠️ DEVOPS TASKS
[Infrastructure updates, deployment script changes, configuration updates]

⚠️ DEPENDENCIES & COORDINATION
[Which tasks depend on each other, handoff points]

🔮 DEPLOYMENT STRATEGY
[How to safely roll out the changes to Hetzner]
```

## Success Metrics
- Each round produces genuinely new insights (not repetition)
- Backend and DevOps perspectives remain distinct and valuable
- Cross-pollination generates insights no single specialist would reach
- Implementation plan is actionable and comprehensive
- Dependencies between backend and devops work are clearly identified
- Deployment strategy minimizes risk and downtime
- Code changes align with infrastructure capabilities
- Infrastructure supports feature requirements

Execute the two-agent analysis with the Backend Engineer and DevOps Engineer working in parallel, then cross-pollinating their findings to produce a coordinated implementation plan.
