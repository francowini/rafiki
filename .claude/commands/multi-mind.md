# Multi-Mind - Rafiki Development Team

Execute a collaborative analysis using specialized subagents to generate a **single deliverables document** with small, implementable tasks for the Rafiki personal development tracker project.

**Usage**: `/multi-mind [topic/task]` or `/multi-mind [topic/task] --no-product` (to exclude Product Agent)

## Project Context

This is a personal development tracking application (Rafiki) with a Go backend, Next.js frontend, and PostgreSQL database. Backend is deployed on Hetzner servers, frontend on Vercel. The application helps users track their ideals (ideales), values (valores), habits (hábitos), goals (metas), and objectives (objetivos).

## Agents

1. **Product Agent** (ALWAYS INCLUDED unless `--no-product`): Mental wellness expert, self-improvement methodology, psychological frameworks
2. **Backend Engineer**: Go development, API design, database operations
3. **UX/UI Designer**: Screen design, shadcn/ui patterns, user experience, visual hierarchy
4. **Frontend Engineer**: Next.js/React development, component implementation, API integration
5. **DevOps Engineer**: Deployment, infrastructure, Docker, server management
6. **Architecture Validator**: Validates proposals against business-model-dependencies.md rules

## Output: Single Deliverables Document

**CRITICAL**: This command generates ONLY specifications, NOT code. The output is a single document:

```
docs/[feature-name]-deliverables.md
```

This document contains small, implementable deliverables that can be executed by:
- `/implement-backend [doc-path] [deliverable-number]`
- `/implement-frontend [doc-path] [deliverable-number]`
- `/implement-devops [doc-path] [deliverable-number]`

## Implementation

Execute analysis using the Task tool to create independent subagents.

**Default: TWO rounds** (can be extended if user requests)

### Round 1: Analysis & Questions

Launch parallel subagents to analyze the topic and ask clarifying questions.

Each agent should:
1. Review the existing codebase (technical agents) or domain context (Product Agent)
2. Understand the requirements
3. **Ask questions to clarify ambiguities or gather requirements**
4. Provide initial analysis based on their specialty

**Agent 0: Product Agent** (ALWAYS RUNS unless `--no-product` specified)
- Prompt: "As a Product Agent specializing in mental wellness, self-improvement, and personal development methodologies, analyze [topic] focusing on:
  - **Psychological frameworks**: Identify relevant mental models, behavioral psychology principles
  - **Self-improvement methodologies**: Consider ACT, SMART goals, habit formation science, mindfulness
  - **User mental wellness**: How does this feature support motivation, well-being, self-efficacy?
  - **User experience from wellness perspective**: Positive reinforcement vs shame-based tracking

  **Ask clarifying questions about**:
  - The psychological intent behind this feature
  - Target user personas and their mental wellness needs
  - How this connects to the broader personal development journey
  - Balance between tracking/metrics and emotional well-being

  Use Read, Glob, and Grep tools to examine existing features."

**Agent 1: Backend Engineer**
- Prompt: "As a backend engineer specializing in Go, analyze [topic] focusing on:
  - Go implementation patterns and best practices
  - PostgreSQL database design
  - API endpoint design (Go 1.22+ patterns)
  - Service architecture (partner-service pattern)
  - Business types pattern (strong types with validation)
  - Review relevant code in api/, foundation/, and business/

  **Ask clarifying questions about**:
  - Required API endpoints and data contracts
  - Database schema decisions
  - Business validation rules

  Use Read, Glob, and Grep tools to examine the existing codebase."

**Agent 2: UX/UI Designer**
- Prompt: "As a UX/UI designer specializing in shadcn/ui, analyze [topic] focusing on:
  - Screen layout and information architecture
  - Visual hierarchy and component composition
  - shadcn/ui component selection (check frontend/components/ui/)
  - Design system consistency (New York style, zinc base, Lucide icons)
  - Responsive design and accessibility
  - Review existing components in frontend/components/features/

  **Ask clarifying questions about**:
  - User preferences for layout and visual style
  - Priority of features for the initial screen
  - Required user interactions and workflows

  Use Read tool to examine existing UI components."

**Agent 3: Frontend Engineer**
- Prompt: "As a frontend engineer specializing in Next.js and React, analyze [topic] focusing on:
  - Next.js 16+ App Router patterns
  - TypeScript type safety
  - React Hook Form with Zod validation
  - shadcn/ui integration
  - API client implementation
  - Review relevant code in frontend/ folder

  **Ask clarifying questions about**:
  - State management approach
  - Error handling preferences
  - Client vs Server component decisions

  Use Read, Glob, and Grep tools to examine the frontend codebase."

**Agent 4: DevOps Engineer**
- Prompt: "As a DevOps engineer, analyze [topic] focusing on:
  - Docker and docker-compose configuration
  - Hetzner server deployment
  - Vercel deployment for frontend
  - Environment variable management
  - Review deployment scripts in devops/ folder

  **Ask clarifying questions about**:
  - Deployment timing and rollout strategy
  - Environment-specific configurations
  - Database migration strategy

  Use Read tool to examine deployment scripts and Docker files."

### Round 2: Synthesis, Architecture Validation & Deliverables Planning

**After user answers Round 1 questions**, launch a second round with all agents plus Architecture Validator.

**Agent 5: Architecture Validator**
- Prompt: "As an Architecture Validator, analyze ALL proposals against `devs/business-model-dependencies.md`:

  **Validate**:
  1. Domain Type Classification (Root/Child/Support/Query)
  2. One-Directional Import Rule (Child → Parent only)
  3. Interface-Based Contracts (ExtBusiness interfaces)
  4. Delegate Pattern for Events
  5. Strong Type Usage (business/types/)
  6. Database as Dumb Storage (no CHECK constraints)

  **Output**:
  ```
  === ARCHITECTURE ALIGNMENT CHECK ===
  Status: [ALIGNED | NOT ALIGNED]
  Violations: [list if any]
  Recommendations: [how to fix]
  ```

  Use Read tool to examine `devs/business-model-dependencies.md`."

### Round 3: Architecture Resolution (if NOT ALIGNED)

If Architecture Validator found issues, present questions to user and re-validate.

### Final Output: Deliverables Document

**CRITICAL**: Generate ONLY `docs/[feature-name]-deliverables.md` with this structure:

```markdown
# [Feature Name] - Implementation Deliverables

## Overview
[Brief description of the feature]

## Architecture Compliance
- Status: ALIGNED
- Domain Type: [Root | Child | Support | Query]
- Parent Domain: [if child domain]

## Product Validation
[Summary of Product Agent's wellness alignment check]

---

## Deliverables

### Deliverable 1: [Short Title]
**Type**: Backend | Frontend | DevOps
**Estimated Scope**: Small | Medium | Large
**Dependencies**: None | Deliverable X

**Description**:
[What needs to be implemented - NO CODE, just specification]

**Acceptance Criteria**:
- [ ] Criterion 1
- [ ] Criterion 2

**Technical Notes**:
- [Key technical consideration]
- [Another consideration]

---

### Deliverable 2: [Short Title]
**Type**: Backend | Frontend | DevOps
...

---

## Deliverable Sequence

Recommended implementation order:
1. Deliverable X (Backend) - foundation
2. Deliverable Y (Backend) - depends on X
3. Deliverable Z (Frontend) - depends on Y
...

## Notes for Implementers

### Backend Notes
[Summary of backend considerations from Backend Engineer]

### Frontend Notes
[Summary of frontend considerations from Frontend Engineer]

### DevOps Notes
[Summary of devops considerations from DevOps Engineer]
```

## Anti-Repetition Mechanisms

**Moderator Responsibilities**:
- Track what has been thoroughly covered
- Redirect specialists away from rehashing previous points
- Ensure each specialist provides distinct value

**Specialist Guidelines**:
- Build on previous round insights rather than restating
- Cross-pollinate by considering other specialists' domain concerns

## Project Phase Context

**IMPORTANT - Project is in MVP phase. Keep recommendations lean:**

- **NO unit tests required** - Focus on shipping features quickly
- **Small, incremental deliverables** - Break work into manageable tasks
- **Minimum viable features** - Build only what's absolutely necessary
- **Simple solutions preferred** - Avoid over-engineering
- **Ship early, iterate fast** - Get features in production quickly

## Success Metrics

- All agents ask meaningful clarifying questions
- User provides clear answers to guide implementation
- Architecture Validator confirms ALIGNED status
- Product Agent validates wellness alignment
- Single deliverables document generated with clear, small tasks
- Each deliverable is independently implementable
- Deliverable sequence is logical and respects dependencies
- **NO CODE in output** - specifications only

Execute the multi-agent analysis in rounds, with user interaction between rounds, and generate a single deliverables document for implementation by the `/implement-*` commands.
