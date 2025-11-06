# Multi-Mind - Rafiki Development Team

Execute a collaborative analysis using three specialized subagents for the Rafiki habits tracker project.

**Usage**: `/multi-mind [topic/task]`

## Project Context

This is a habits tracker application (Rafiki) with a Go backend, Next.js frontend, and PostgreSQL database. Backend is deployed on Hetzner servers, frontend on Vercel. The three-agent team consists of:
1. **Backend Engineer**: Go development, API design, database operations
2. **Frontend Engineer**: Next.js/React development, UI/UX, API integration
3. **DevOps Engineer**: Deployment, infrastructure, Docker, server management

## Implementation

Execute this three-specialist analysis using the Task tool to create independent subagents:

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
  - CORS configuration for frontend integration
  - Review relevant code in api/services/partner/, foundation/, and business/
  Use Read, Glob, and Grep tools to examine the existing codebase."

**Agent 2: Frontend Engineer**
- Prompt: "As a frontend engineer specializing in Next.js and React, analyze [topic] focusing on:
  - Next.js 14+ App Router patterns and best practices
  - TypeScript type safety and interface design
  - shadcn/ui component usage and customization
  - Tailwind CSS styling and responsive design
  - API client implementation and error handling
  - State management with React hooks
  - Form validation and user experience
  - Performance optimization (loading states, pagination)
  - Vercel deployment considerations
  - Review relevant code in frontend/ folder (app/, components/, lib/)
  Use Read, Glob, and Grep tools to examine the frontend codebase."

**Agent 3: DevOps Engineer**
- Prompt: "As a DevOps engineer, analyze [topic] focusing on:
  - Docker and docker-compose configuration for backend
  - Hetzner server deployment strategies
  - Vercel deployment and configuration for frontend
  - Service deployment and orchestration
  - Environment variable management (both backend and frontend)
  - Health checks and monitoring
  - CORS and network configuration
  - SSL/TLS certificate management
  - Nginx reverse proxy configuration
  - Review deployment scripts and documentation in devops/ folder
  Use Read tool to examine deployment scripts, Docker files, and configuration files."

Each subagent operates independently with access to Read, Glob, Grep, and analysis tools.

**Specialist Domains**:
- **Backend**: Go codebase, API logic, database schemas, business logic, CORS
- **Frontend**: Next.js application, React components, UI/UX, TypeScript, API integration
- **DevOps**: Infrastructure (Hetzner + Vercel), deployment scripts, containerization, server configuration

### Phase 2: Cross-Pollination Round

After receiving all three specialist reports, launch a second round of subagents:

```
Backend Engineer Review:
- Review Frontend findings for API contract considerations
- Review DevOps findings for deployment considerations
- Identify how infrastructure choices affect code design
- Ensure API responses match frontend expectations
- Ensure CORS configuration supports frontend needs
- Flag potential issues with containerization or environment setup

Frontend Engineer Review:
- Review Backend findings for API integration considerations
- Review DevOps findings for deployment and environment configuration
- Identify API changes needed for better UX
- Ensure error handling covers all API error cases
- Verify environment variable usage aligns with Vercel deployment
- Flag potential performance or user experience concerns

DevOps Engineer Review:
- Review Backend findings for deployment implications
- Review Frontend findings for Vercel deployment requirements
- Identify infrastructure requirements for new features
- Ensure deployment scripts support both backend and frontend changes
- Verify CORS, SSL, and networking configuration supports full stack
- Flag potential performance or scaling concerns
```

### Phase 3: Synthesis & Final Recommendations

After both rounds:
- Collect all subagent outputs
- Synthesize backend, frontend, and devops perspectives
- Identify coordination points between frontend, backend, and infrastructure
- Provide actionable implementation plan
- Highlight dependencies between backend, frontend, and devops work
- Ensure API contracts are clear and well-documented

## Anti-Repetition Mechanisms

**Moderator Responsibilities**:
- Track what has been thoroughly covered vs. what needs deeper exploration
- Redirect specialists away from rehashing previous points
- Push for new angles, deeper analysis, or broader implications
- Ensure backend, frontend, and devops perspectives remain distinct and valuable

**Specialist Guidelines**:
- Build on previous round insights rather than restating them
- Backend focuses on code quality, maintainability, Go best practices, and API contracts
- Frontend focuses on user experience, TypeScript safety, component design, and API integration
- DevOps focuses on deployability, reliability, operational excellence, and cross-platform configuration
- Cross-pollinate by considering the other specialists' domain concerns

## Output Protocol

```
=== RAFIKI HABITS TRACKER ANALYSIS: [Topic] ===
Specialists: Backend Engineer + Frontend Engineer + DevOps Engineer

--- ROUND 1: INITIAL ANALYSIS ---
🔧 BACKEND ENGINEER ANALYSIS
[Go code analysis, API design, database considerations, CORS configuration]

🎨 FRONTEND ENGINEER ANALYSIS
[Next.js/React analysis, UI/UX considerations, API integration, TypeScript types]

🚀 DEVOPS ENGINEER ANALYSIS
[Infrastructure analysis, deployment strategy (Hetzner + Vercel), operational concerns]

--- ROUND 2: CROSS-POLLINATION ---
🔧 BACKEND ENGINEER REVIEW
[Review of Frontend and DevOps findings and their impact on API design]

🎨 FRONTEND ENGINEER REVIEW
[Review of Backend and DevOps findings and their impact on UI/UX and integration]

🚀 DEVOPS ENGINEER REVIEW
[Review of Backend and Frontend findings and their deployment implications]

--- FINAL SYNTHESIS ---
🎯 IMPLEMENTATION PLAN
[Step-by-step plan coordinating backend, frontend, and devops work]

💻 BACKEND TASKS
[Specific Go code changes, database migrations, API updates, CORS configuration]

🎨 FRONTEND TASKS
[Component updates, API client changes, UI enhancements, type definitions]

🛠️ DEVOPS TASKS
[Infrastructure updates, deployment script changes, configuration updates for both platforms]

⚠️ DEPENDENCIES & COORDINATION
[Which tasks depend on each other, handoff points, API contract agreements]

🔮 DEPLOYMENT STRATEGY
[How to safely roll out the changes to Hetzner (backend) and Vercel (frontend)]
```

## Success Metrics
- Each round produces genuinely new insights (not repetition)
- Backend, Frontend, and DevOps perspectives remain distinct and valuable
- Cross-pollination generates insights no single specialist would reach
- Implementation plan is actionable and comprehensive
- Dependencies between backend, frontend, and devops work are clearly identified
- Deployment strategy minimizes risk and downtime
- Code changes align with infrastructure capabilities
- Infrastructure supports feature requirements
- API contracts are clear and well-documented
- Frontend UX is smooth and error handling is comprehensive
- CORS and networking configuration supports full-stack integration

Execute the three-agent analysis with the Backend Engineer, Frontend Engineer, and DevOps Engineer working in parallel, then cross-pollinating their findings to produce a coordinated implementation plan.
